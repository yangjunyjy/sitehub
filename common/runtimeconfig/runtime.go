// Package runtimeconfig publishes immutable, validated traffic settings.
package runtimeconfig

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Config struct {
	Maintenance           bool  `json:",default=false"`
	MaxConcurrentRequests int64 `json:",default=0"`
}

type Snapshot struct {
	Config  Config
	Version string
}
type Status struct {
	AppliedVersion string    `json:"appliedVersion"`
	DesiredVersion string    `json:"desiredVersion"`
	Source         string    `json:"source"`
	LastError      string    `json:"lastError,omitempty"`
	AppliedAt      time.Time `json:"appliedAt"`
}
type Manager struct {
	current atomic.Pointer[Snapshot]
	active  atomic.Int64
	mu      sync.Mutex
	status  Status
}

func Version(data []byte) string { sum := sha256.Sum256(data); return hex.EncodeToString(sum[:]) }
func New(c Config, version, source string) (*Manager, error) {
	m := &Manager{status: Status{Source: source}}
	if err := m.Commit(c, version); err != nil {
		return nil, err
	}
	return m, nil
}
func (m *Manager) Commit(c Config, version string) error {
	if c.MaxConcurrentRequests < 0 || c.MaxConcurrentRequests > 1000000 {
		return fmt.Errorf("Runtime.MaxConcurrentRequests must be between 0 and 1000000")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.current.Store(&Snapshot{Config: c, Version: version})
	m.status.AppliedVersion = version
	m.status.DesiredVersion = version
	m.status.LastError = ""
	m.status.AppliedAt = time.Now().UTC()
	return nil
}
func (m *Manager) Reject(version, reason string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.status.DesiredVersion = version
	m.status.LastError = reason
}
func (m *Manager) Status() Status     { m.mu.Lock(); defer m.mu.Unlock(); return m.status }
func (m *Manager) Snapshot() Snapshot { return *m.current.Load() }

// Each request reads exactly one immutable snapshot. Lowering the limit does
// not interrupt work already admitted, including downloads and RPC streams.
func (m *Manager) Enter() (func(), error) {
	c := m.Snapshot().Config
	if c.Maintenance {
		return nil, status.Error(codes.Unavailable, "service maintenance")
	}
	n := m.active.Add(1)
	if c.MaxConcurrentRequests > 0 && n > c.MaxConcurrentRequests {
		m.active.Add(-1)
		return nil, status.Error(codes.ResourceExhausted, "service concurrency limit")
	}
	return func() { m.active.Add(-1) }, nil
}
func (m *Manager) HTTP(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		release, err := m.Enter()
		if err != nil {
			code := http.StatusServiceUnavailable
			if status.Code(err) == codes.ResourceExhausted {
				code = http.StatusTooManyRequests
			}
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(code)
			_ = json.NewEncoder(w).Encode(map[string]any{"code": code, "message": status.Convert(err).Message()})
			return
		}
		defer release()
		next(w, r)
	}
}
func (m *Manager) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		release, err := m.Enter()
		if err != nil {
			return nil, err
		}
		defer release()
		return handler(ctx, req)
	}
}
func (m *Manager) Stream() grpc.StreamServerInterceptor {
	return func(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		release, err := m.Enter()
		if err != nil {
			return err
		}
		defer release()
		return handler(srv, stream)
	}
}
