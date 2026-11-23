package engine

import (
	"context"
	"io"
	"net"
	"testing"
	"time"

	"sitehub/gateway/manifest"
	"sitehub/gateway/plugin"
)

func TestStreamDrainHonorsExistingTCPAndUDPPeers(t *testing.T) {
	tcp := tcpBackend(t, "T")
	udp, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer udp.Close()
	go func() {
		buf := make([]byte, 64)
		for {
			n, src, err := udp.ReadFrom(buf)
			if err != nil {
				return
			}
			_, _ = udp.WriteTo(buf[:n], src)
		}
	}()
	tcpTemp, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	tcpAddr := tcpTemp.Addr().String()
	tcpTemp.Close()
	udpTemp, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	udpAddr := udpTemp.LocalAddr().String()
	udpTemp.Close()
	e := New(nil, plugin.DefaultRegistry(), 1024)
	defer e.Close()
	commit(t, e, manifest.Document{Version: 1, Streams: []manifest.Stream{
		{ID: "tcp", Protocol: "tcp", Listen: tcpAddr, Target: tcp.Addr().String()},
		{ID: "udp", Protocol: "udp", Listen: udpAddr, Target: udp.LocalAddr().String()},
	}})

	dial := func(protocol, addr string) net.Conn {
		t.Helper()
		c, err := net.DialTimeout(protocol, addr, time.Second)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { c.Close() })
		_ = c.SetDeadline(time.Now().Add(4 * time.Second))
		return c
	}
	tcpClient := dial("tcp", tcpAddr).(*net.TCPConn)
	if _, err := io.ReadFull(tcpClient, make([]byte, 1)); err != nil {
		t.Fatal(err)
	}
	udpClient := dial("udp", udpAddr)
	echo := func(c net.Conn) {
		t.Helper()
		if _, err := c.Write([]byte("payload")); err != nil {
			t.Fatal(err)
		}
		buf := make([]byte, 7)
		if _, err := io.ReadFull(c, buf); err != nil || string(buf) != "payload" {
			t.Fatalf("echo %q: %v", buf, err)
		}
	}
	echo(udpClient)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	go func() { e.DrainStreams(ctx); close(done) }()
	eventually(t, "draining", func() bool { e.mu.Lock(); defer e.mu.Unlock(); return e.streamDraining })
	if fresh, err := net.DialTimeout("tcp", tcpAddr, 200*time.Millisecond); err == nil {
		fresh.Close()
		t.Fatal("new TCP flow admitted during drain")
	}
	echo(udpClient)
	newPeer := dial("udp", udpAddr)
	newPeer.SetReadDeadline(time.Now().Add(80 * time.Millisecond))
	newPeer.Write([]byte("new"))
	if _, err := newPeer.Read(make([]byte, 64)); err == nil {
		t.Fatal("new UDP association admitted during drain")
	}
	tcpClient.Write([]byte("finish"))
	tcpClient.CloseWrite()
	if b, err := io.ReadAll(tcpClient); err != nil || string(b) != "finish" {
		t.Fatalf("TCP drain lost %q: %v", b, err)
	}
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("stream shutdown missed deadline")
	}
}
