package engine

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"net/http"
	"time"
)

// Each successful I/O operation renews the idle budget. A continuously active
// upload, download or SSE response is not cut off by a total request timeout.
type idleConn struct {
	net.Conn
	timeout time.Duration
}

func (c *idleConn) Read(p []byte) (int, error) {
	_ = c.Conn.SetReadDeadline(time.Now().Add(c.timeout))
	return c.Conn.Read(p)
}
func (c *idleConn) Write(p []byte) (int, error) {
	_ = c.Conn.SetWriteDeadline(time.Now().Add(c.timeout))
	return c.Conn.Write(p)
}

type idleBody struct {
	io.ReadCloser
	control *http.ResponseController
	timeout time.Duration
}

func (b *idleBody) Read(p []byte) (int, error) {
	_ = b.control.SetReadDeadline(time.Now().Add(b.timeout))
	n, err := b.ReadCloser.Read(p)
	if err != nil {
		_ = b.control.SetReadDeadline(time.Time{})
	}
	return n, err
}
func (b *idleBody) Close() error {
	_ = b.control.SetReadDeadline(time.Time{})
	return b.ReadCloser.Close()
}

type idleWriter struct {
	http.ResponseWriter
	control *http.ResponseController
	timeout time.Duration
}

func (w *idleWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }
func (w *idleWriter) Write(p []byte) (int, error) {
	_ = w.control.SetWriteDeadline(time.Now().Add(w.timeout))
	return w.ResponseWriter.Write(p)
}
func (w *idleWriter) WriteHeader(status int) {
	_ = w.control.SetWriteDeadline(time.Now().Add(w.timeout))
	w.ResponseWriter.WriteHeader(status)
}
func (w *idleWriter) Flush() {
	_ = w.control.SetWriteDeadline(time.Now().Add(w.timeout))
	_ = w.control.Flush()
}

// Upgraded connections bypass ResponseWriter. Carry the idle budget onto the
// client socket as well, preserving data already read ahead by net/http.
func (w *idleWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	conn, buffered, err := w.control.Hijack()
	if err != nil {
		return nil, nil, err
	}
	wrapped := &idleConn{Conn: conn, timeout: w.timeout}
	if err = buffered.Writer.Flush(); err != nil {
		_ = wrapped.Close()
		return nil, nil, err
	}
	var pending []byte
	if n := buffered.Reader.Buffered(); n > 0 {
		data, _ := buffered.Reader.Peek(n)
		pending = append([]byte(nil), data...)
	}
	client := &prefixedConn{Conn: wrapped, reader: io.MultiReader(bytes.NewReader(pending), wrapped)}
	return client, bufio.NewReadWriter(bufio.NewReader(client), bufio.NewWriter(client)), nil
}

type prefixedConn struct {
	net.Conn
	reader io.Reader
}

func (c *prefixedConn) Read(p []byte) (int, error) { return c.reader.Read(p) }
