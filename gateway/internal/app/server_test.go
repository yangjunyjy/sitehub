package app

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"
)

func trackedServer(t *testing.T, secure bool, handler http.Handler) (*httptest.Server, *connections) {
	t.Helper()
	server := httptest.NewUnstartedServer(handler)
	connections := newConnections()
	server.Listener = &trackingListener{Listener: server.Listener, owner: connections}
	if secure {
		server.StartTLS()
	} else {
		server.Start()
	}
	t.Cleanup(func() { connections.closeAll(); server.Close() })
	return server, connections
}

func openUpgrade(t *testing.T, server *httptest.Server) net.Conn {
	t.Helper()
	u, _ := url.Parse(server.URL)
	var conn net.Conn
	var err error
	if u.Scheme == "https" {
		// Use httptest's per-server trusted CA instead of bypassing validation.
		config := server.Client().Transport.(*http.Transport).TLSClientConfig.Clone()
		conn, err = tls.DialWithDialer(&net.Dialer{Timeout: 3 * time.Second}, "tcp", u.Host, config)
	} else {
		conn, err = net.DialTimeout("tcp", u.Host, 3*time.Second)
	}
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
	_, _ = fmt.Fprintf(conn, "GET /upgrade HTTP/1.1\r\nHost: %s\r\nConnection: Upgrade\r\nUpgrade: websocket\r\n\r\n", u.Host)
	response, err := http.ReadResponse(bufio.NewReader(conn), &http.Request{Method: "GET"})
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("unexpected upgrade status: %d", response.StatusCode)
	}
	return conn
}

func upgrade(w http.ResponseWriter) (net.Conn, *bufio.ReadWriter, error) {
	conn, rw, err := http.NewResponseController(w).Hijack()
	if err != nil {
		return nil, nil, err
	}
	_, _ = rw.WriteString("HTTP/1.1 101 Switching Protocols\r\nConnection: Upgrade\r\nUpgrade: websocket\r\n\r\n")
	_ = rw.Flush()
	return conn, rw, nil
}

func activeCount(c *connections) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.active)
}

func TestClosedUpgradedConnectionDoesNotLeakOrDelayDrain(t *testing.T) {
	for _, secure := range []bool{false, true} {
		t.Run(fmt.Sprintf("TLS=%v", secure), func(t *testing.T) {
			closed := make(chan struct{})
			server, connections := trackedServer(t, secure, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				conn, _, err := upgrade(w)
				if err != nil {
					t.Error(err)
					close(closed)
					return
				}
				_ = conn.Close()
				close(closed)
			}))
			openUpgrade(t, server)
			select {
			case <-closed:
			case <-time.After(3 * time.Second):
				t.Fatal("upgraded handler did not close")
			}
			if activeCount(connections) != 0 {
				t.Fatal("closed upgraded connection retained")
			}
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			start := time.Now()
			connections.drain(ctx)
			if elapsed := time.Since(start); elapsed > 250*time.Millisecond {
				t.Fatalf("empty connection tracker delayed shutdown: %s", elapsed)
			}
		})
	}
}

func TestActiveUpgradeClosesAtSharedDeadline(t *testing.T) {
	for _, secure := range []bool{false, true} {
		t.Run(fmt.Sprintf("TLS=%v", secure), func(t *testing.T) {
			finished := make(chan struct{})
			server, connections := trackedServer(t, secure, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer close(finished)
				conn, rw, err := upgrade(w)
				if err != nil {
					t.Error(err)
					return
				}
				defer conn.Close()
				_, _ = rw.ReadByte()
			}))
			client := openUpgrade(t, server)
			if activeCount(connections) != 1 {
				t.Fatal("active upgrade not tracked")
			}
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()
			if err := server.Config.Shutdown(ctx); err != nil {
				t.Fatal(err)
			}
			connections.drain(ctx)
			if ctx.Err() == nil || activeCount(connections) != 0 {
				t.Fatal("upgrade was not drained at its deadline")
			}
			var value [1]byte
			if _, err := client.Read(value[:]); err == nil {
				t.Fatal("upgraded client socket remained open")
			}
			select {
			case <-finished:
			case <-time.After(time.Second):
				t.Fatal("upgraded handler remained blocked after socket close")
			}
		})
	}
}

func TestUpgradeNaturalCloseWakesDrain(t *testing.T) {
	server, connections := trackedServer(t, false, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, rw, err := upgrade(w)
		if err != nil {
			t.Error(err)
			return
		}
		defer conn.Close()
		_, _ = rw.ReadByte()
	}))
	client := openUpgrade(t, server)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := server.Config.Shutdown(ctx); err != nil {
		t.Fatal(err)
	}
	drained := make(chan struct{})
	go func() { connections.drain(ctx); close(drained) }()
	select {
	case <-drained:
		t.Fatal("active socket was not awaited")
	case <-time.After(20 * time.Millisecond):
	}
	_ = client.Close()
	select {
	case <-drained:
		if ctx.Err() != nil || activeCount(connections) != 0 {
			t.Fatal("natural close did not finish before deadline")
		}
	case <-time.After(time.Second):
		t.Fatal("natural socket close did not wake drain")
	}
}

func TestGracefulShutdownPreservesSlowHTTPResponse(t *testing.T) {
	started, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	unblock := func() { once.Do(func() { close(release) }) }
	defer unblock()
	server, connections := trackedServer(t, false, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("first-"))
		_ = http.NewResponseController(w).Flush()
		close(started)
		<-release
		_, _ = w.Write([]byte("last"))
	}))
	response, err := server.Client().Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	<-started
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	finished := make(chan error, 1)
	go func() {
		err := server.Config.Shutdown(ctx)
		connections.drain(ctx)
		finished <- err
	}()
	select {
	case <-finished:
		t.Fatal("shutdown cut an in-flight HTTP response")
	case <-time.After(30 * time.Millisecond):
	}
	unblock()
	data, err := io.ReadAll(response.Body)
	if err != nil || strings.TrimSpace(string(data)) != "first-last" {
		t.Fatalf("response was truncated: %q, %v", data, err)
	}
	if err := <-finished; err != nil || ctx.Err() != nil || activeCount(connections) != 0 {
		t.Fatalf("graceful shutdown failed: %v", err)
	}
}
