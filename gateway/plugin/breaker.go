package plugin

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type breakerConfig struct {
	Failures           int `json:"failures"`
	Successes          int `json:"successes"`
	InitialOpenSeconds int `json:"initial_open_seconds"`
	MaxOpenSeconds     int `json:"max_open_seconds"`
}

type breaker struct {
	mu                  sync.Mutex
	config              breakerConfig
	state               string
	generation          uint64
	failures, successes int
	backoff             time.Duration
	until               time.Time
	probe               bool
}

func circuitBreaker(raw json.RawMessage) (Middleware, error) {
	c := breakerConfig{Failures: 5, Successes: 2, InitialOpenSeconds: 1, MaxOpenSeconds: 30}
	if err := decode(raw, &c); err != nil {
		return nil, err
	}
	if c.Failures < 1 || c.Failures > 100000 || c.Successes < 1 || c.Successes > 100000 || c.InitialOpenSeconds < 1 || c.MaxOpenSeconds < c.InitialOpenSeconds || c.MaxOpenSeconds > 3600 {
		return nil, fmt.Errorf("positive bounded failures/successes/open durations required")
	}
	state := &breaker{config: c, state: "closed"}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			generation, allowed, retry := state.allow(time.Now())
			if !allowed {
				w.Header().Set("Retry-After", strconv.Itoa(retry))
				respond(w, r, http.StatusServiceUnavailable, "Upstream temporarily unavailable")
				return
			}
			writer := &statusWriter{ResponseWriter: w}
			completed := false
			defer func() {
				status := writer.status
				if !completed {
					status = http.StatusBadGateway
				} else if status == 0 {
					status = http.StatusOK
				}
				if r.Context().Err() != nil {
					status = 0
				}
				state.finish(generation, status, time.Now())
			}()
			next.ServeHTTP(writer, r)
			completed = true
		})
	}, nil
}

func (b *breaker) allow(now time.Time) (uint64, bool, int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.state == "open" {
		if now.Before(b.until) {
			return b.generation, false, max(1, int(b.until.Sub(now)/time.Second)+1)
		}
		b.state, b.probe, b.successes = "half-open", false, 0
		b.generation++
	}
	if b.state == "half-open" {
		if b.probe {
			return b.generation, false, 1
		}
		b.probe = true
	}
	return b.generation, true, 0
}

func (b *breaker) finish(generation uint64, status int, now time.Time) {
	b.mu.Lock()
	defer b.mu.Unlock()
	// Responses already in flight before an open/close transition cannot
	// overwrite the new state's counters, even when they finish much later.
	if generation != b.generation {
		return
	}
	failed := status == 500 || status == 502 || status == 503 || status == 504
	healthy := status == 101 || status == 200 || status == 201 || status == 202 || status == 204 || status == 206 || status == 304
	if b.state == "half-open" {
		b.probe = false
		if failed {
			b.open(now)
			return
		}
		if healthy {
			b.successes++
			if b.successes >= b.config.Successes {
				b.state, b.failures, b.successes, b.backoff = "closed", 0, 0, 0
				b.generation++
			}
		}
		return
	}
	if b.state != "closed" {
		return
	}
	if healthy {
		b.failures = 0
	}
	if failed {
		b.failures++
		if b.failures >= b.config.Failures {
			b.open(now)
		}
	}
}

func (b *breaker) open(now time.Time) {
	if b.backoff == 0 {
		b.backoff = time.Duration(b.config.InitialOpenSeconds) * time.Second
	} else {
		b.backoff *= 2
	}
	b.backoff = min(b.backoff, time.Duration(b.config.MaxOpenSeconds)*time.Second)
	b.state, b.until, b.probe, b.successes = "open", now.Add(b.backoff), false, 0
	b.generation++
}

// statusWriter observes headers only. It never buffers bodies, and delegates
// ResponseController operations so ReverseProxy can stream or upgrade sockets.
type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }
func (w *statusWriter) WriteHeader(status int) {
	if w.status != 0 {
		return
	}
	if status >= 200 || status == 101 {
		w.status = status
	}
	w.ResponseWriter.WriteHeader(status)
}
func (w *statusWriter) Write(data []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(data)
}
func (w *statusWriter) FlushError() error {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	return http.NewResponseController(w.ResponseWriter).Flush()
}
func (w *statusWriter) Flush() { _ = w.FlushError() }
func (w *statusWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	conn, rw, err := http.NewResponseController(w.ResponseWriter).Hijack()
	if err == nil && w.status == 0 {
		w.status = http.StatusSwitchingProtocols
	}
	return conn, rw, err
}
