package engine

import (
	"bufio"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type upgradeWriter struct {
	*httptest.ResponseRecorder
	conn net.Conn
}

func (w *upgradeWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.conn, bufio.NewReadWriter(bufio.NewReader(w.conn), bufio.NewWriter(w.conn)), nil
}

func TestUpgradedClientBackpressureHasIdleDeadline(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()
	w := &upgradeWriter{ResponseRecorder: httptest.NewRecorder(), conn: server}
	idle := &idleWriter{ResponseWriter: w, control: http.NewResponseController(w), timeout: 50 * time.Millisecond}
	conn, _, err := idle.Hijack()
	if err != nil {
		t.Fatal(err)
	}
	// A peer that never reads must not leave a proxy copy goroutine blocked
	// forever after the response has upgraded and bypassed ResponseWriter.
	done := make(chan error, 1)
	go func() { _, err := conn.Write([]byte("unread frame")); done <- err }()
	select {
	case err := <-done:
		if timeout, ok := err.(net.Error); !ok || !timeout.Timeout() {
			t.Fatalf("expected idle write timeout, got %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("upgraded client write has no idle deadline")
	}
}
