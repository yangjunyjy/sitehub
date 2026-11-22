package engine

import (
	"context"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"
)

const udpIdleTimeout = 30 * time.Second

var streamDialer = net.Dialer{Timeout: 5 * time.Second}

type streamSocket struct {
	engine   *Engine
	key      string
	listener net.Listener
	packet   net.PacketConn
	refs     int
	started  bool
}

// Reserve all sockets before CAS, without accepting traffic until commit. Shared
// sockets read their target from the same atomic snapshot as HTTP requests.
func (e *Engine) reserveStreams(s *Snapshot) error {
	s.streamTargets = make(map[*streamSocket]string, len(s.Document.Streams))
	for _, rule := range s.Document.Streams {
		key := rule.Protocol + "/" + strings.ToLower(rule.Listen)
		socket := e.streamSockets[key]
		if socket == nil {
			socket = &streamSocket{engine: e, key: key}
			var err error
			if rule.Protocol == "tcp" {
				socket.listener, err = net.Listen("tcp", rule.Listen)
			} else {
				socket.packet, err = net.ListenPacket("udp", rule.Listen)
			}
			if err != nil {
				return fmt.Errorf("stream %s: bind %s: %w", rule.ID, key, err)
			}
			e.streamSockets[key] = socket
		}
		socket.refs++
		s.heldSockets = append(s.heldSockets, socket)
		s.streamTargets[socket] = rule.Target
	}
	return nil
}

func (e *Engine) releaseStreams(s *Snapshot) {
	for _, socket := range s.heldSockets {
		socket.refs--
		if socket.refs == 0 {
			delete(e.streamSockets, socket.key)
			socket.close()
		}
	}
	s.heldSockets = nil
}

func (s *streamSocket) close() {
	if s.listener != nil {
		s.listener.Close()
	} else {
		s.packet.Close()
	}
}

func (s *streamSocket) target() string {
	if snapshot := s.engine.current.Load(); snapshot != nil {
		return snapshot.streamTargets[s]
	}
	return ""
}

func (s *streamSocket) serve() {
	defer s.engine.streamWorkers.Done()
	if s.packet != nil {
		s.servePackets()
		return
	}
	for {
		client, err := s.listener.Accept()
		if err != nil {
			return
		}
		target := s.target()
		if target == "" || !s.engine.startFlow() {
			client.Close()
			continue
		}
		go s.engine.proxyTCP(client.(*net.TCPConn), target)
	}
}

func (e *Engine) startFlow() bool {
	select {
	case e.streamSlots <- struct{}{}:
		e.streamWorkers.Add(1)
		return true
	default:
		return false
	}
}

func (e *Engine) finishFlow() {
	<-e.streamSlots
	e.streamWorkers.Done()
}

// Both directions stay open independently so request/response protocols using
// FIN as the end of a request can still read the complete upstream response.
func (e *Engine) proxyTCP(client *net.TCPConn, target string) {
	defer e.finishFlow()
	defer client.Close()
	conn, err := streamDialer.DialContext(e.streamContext, "tcp", target)
	if err != nil {
		return
	}
	upstream := conn.(*net.TCPConn)
	defer upstream.Close()
	closeBoth := func() {
		client.Close()
		upstream.Close()
	}
	stop := context.AfterFunc(e.streamContext, closeBoth)
	defer stop()
	copyDirection := func(dst, src *net.TCPConn) {
		if _, err := io.Copy(dst, src); err != nil {
			closeBoth()
		} else {
			dst.CloseWrite()
		}
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		copyDirection(upstream, client)
	}()
	copyDirection(client, upstream)
	<-done
}

// A source address and port share one upstream socket for the lifetime of the
// UDP conversation. Separate queues keep DNS and upstream I/O off the listener.
func (s *streamSocket) servePackets() {
	ctx, cancel := context.WithCancel(s.engine.streamContext)
	defer cancel()
	var mu sync.Mutex
	peers := map[string]chan []byte{}
	buf := make([]byte, 65535)
	for {
		n, client, err := s.packet.ReadFrom(buf)
		if err != nil {
			return
		}
		target := s.target()
		if target == "" {
			continue
		}
		key := client.String()
		mu.Lock()
		packets := peers[key]
		if packets == nil && s.engine.startFlow() {
			packets = make(chan []byte, 8)
			peers[key] = packets
			go func(input <-chan []byte) {
				defer s.engine.finishFlow()
				s.proxyUDP(ctx, client, target, input)
				mu.Lock()
				delete(peers, key)
				mu.Unlock()
			}(packets)
		}
		if packets != nil && len(packets) < cap(packets) {
			packets <- append([]byte(nil), buf[:n]...)
		}
		mu.Unlock()
	}
}

func (s *streamSocket) proxyUDP(ctx context.Context, client net.Addr, target string, packets <-chan []byte) {
	upstream, err := streamDialer.DialContext(ctx, "udp", target)
	if err != nil {
		return
	}
	defer upstream.Close()
	upstream.SetReadDeadline(time.Now().Add(udpIdleTimeout))
	done := make(chan struct{})
	go func() {
		defer close(done)
		buf := make([]byte, 65535)
		for {
			n, err := upstream.Read(buf)
			if err != nil {
				return
			}
			upstream.SetReadDeadline(time.Now().Add(udpIdleTimeout))
			if _, err := s.packet.WriteTo(buf[:n], client); err != nil {
				return
			}
		}
	}()
	defer func() {
		upstream.Close()
		<-done
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case <-done:
			return
		case datagram := <-packets:
			upstream.SetDeadline(time.Now().Add(udpIdleTimeout))
			// WriteMsgUDP also sends empty datagrams on Windows.
			if _, _, err := upstream.(*net.UDPConn).WriteMsgUDP(datagram, nil, nil); err != nil {
				return
			}
		}
	}
}
