package engine

import (
	"bytes"
	"io"
	"net"
	"testing"
	"time"

	"sitehub/gateway/manifest"
	"sitehub/gateway/plugin"
)

func tcpBackend(t *testing.T, greeting string) net.Listener {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ln.Close() })
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func() {
				defer conn.Close()
				io.WriteString(conn, greeting)
				request, _ := io.ReadAll(conn)
				conn.Write(request)
			}()
		}
	}()
	return ln
}

func TestTCPStreamReloadKeepsConnectionsAndHonorsHalfClose(t *testing.T) {
	first, second := tcpBackend(t, "A"), tcpBackend(t, "B")
	reserved, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := reserved.Addr().String()
	reserved.Close()
	doc := manifest.Document{Version: 1, Streams: []manifest.Stream{{ID: "svc", Listen: address, Protocol: "tcp", Target: first.Addr().String()}}}
	e := New(nil, plugin.DefaultRegistry(), 1024)
	defer e.Close()
	commit(t, e, doc)
	dial := func(want string) *net.TCPConn {
		t.Helper()
		conn, err := net.DialTimeout("tcp", address, time.Second)
		if err != nil {
			t.Fatal(err)
		}
		conn.SetDeadline(time.Now().Add(3 * time.Second))
		got := make([]byte, len(want))
		if _, err := io.ReadFull(conn, got); err != nil || string(got) != want {
			t.Fatalf("greeting %q: %v", got, err)
		}
		t.Cleanup(func() { conn.Close() })
		return conn.(*net.TCPConn)
	}
	old := dial("A")
	doc.Streams[0].Target = second.Addr().String()
	commit(t, e, doc)
	dial("B").Close()
	doc.Streams = append(doc.Streams, manifest.Stream{ID: "conflict", Protocol: "tcp", Listen: second.Addr().String(), Target: address})
	if candidate, err := e.Prepare(doc); err == nil {
		candidate.Release()
		t.Fatal("accepted occupied TCP listener")
	}
	dial("B").Close()
	old.Write([]byte("request"))
	old.CloseWrite()
	if got, err := io.ReadAll(old); err != nil || string(got) != "request" {
		t.Fatalf("half-closed existing connection: %q, %v", got, err)
	}
}

func TestUDPPreservesPacketBoundariesAndClientAssociations(t *testing.T) {
	echo, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer echo.Close()
	go func() {
		buf := make([]byte, 65535)
		for {
			n, peer, err := echo.ReadFrom(buf)
			if err != nil {
				return
			}
			echo.WriteTo([]byte(peer.String()), peer)
			echo.WriteTo(buf[:n], peer)
		}
	}()
	temp, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := temp.LocalAddr().String()
	temp.Close()
	e := New(nil, plugin.DefaultRegistry(), 1024)
	defer e.Close()
	commit(t, e, manifest.Document{Version: 1, Streams: []manifest.Stream{{ID: "udp", Protocol: "udp", Listen: address, Target: echo.LocalAddr().String()}}})
	seenPeers := map[string]bool{}
	for i := 0; i < 2; i++ {
		conn, err := net.Dial("udp", address)
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()
		conn.SetDeadline(time.Now().Add(3 * time.Second))
		var peer string
		for _, payload := range [][]byte{[]byte("hello"), {}, bytes.Repeat([]byte("b"), 1600)} {
			if _, _, err := conn.(*net.UDPConn).WriteMsgUDP(payload, nil, nil); err != nil {
				t.Fatal(err)
			}
			buf := make([]byte, 2048)
			n, err := conn.Read(buf)
			if err != nil {
				t.Fatal(err)
			}
			if peer != "" && string(buf[:n]) != peer {
				t.Fatal("changed association")
			}
			peer = string(buf[:n])
			n, err = conn.Read(buf)
			if err != nil || !bytes.Equal(buf[:n], payload) {
				t.Fatalf("bad echo length %d: %v", n, err)
			}
		}
		if seenPeers[peer] {
			t.Fatal("multiple clients shared one association")
		}
		seenPeers[peer] = true
	}
}
