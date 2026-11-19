// Package registry defines provider-independent service registration and discovery.
package registry

import (
	"context"
	"time"
)

type Config struct {
	Enabled        bool          `json:",default=false"`
	AdvertiseIP    string        `json:",optional"`
	AllowLoopback  bool          `json:",default=false"`
	MaxStale       time.Duration `json:",default=30s"`
	StartupTimeout time.Duration `json:",default=30s"`
}

// Instance is the wire-independent service address published by a provider.
type Instance struct {
	ID       string            `json:"id"`
	Service  string            `json:"service"`
	IP       string            `json:"ip"`
	Port     uint64            `json:"port"`
	Weight   float64           `json:"weight"`
	Healthy  bool              `json:"healthy"`
	Enabled  bool              `json:"enabled"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

type Registration interface{ Close() error }

// Discovery delivers complete snapshots. An empty successful snapshot removes
// every instance; a failed update must not be interpreted as an empty snapshot.
// Cancel is idempotent and may be called from the update callback itself.
type Discovery interface {
	Watch(context.Context, string, func([]Instance, error)) (func(), error)
	List(context.Context, string) ([]Instance, error)
	Healthy() bool
}

type Backend interface {
	Discovery
	Register(context.Context, Instance) (Registration, error)
	Close() error
}

// Proxy keeps consumers bound to this contract rather than a provider SDK.
// The owner closes the proxy after its consumers have stopped.
type Proxy struct{ backend Backend }

func NewProxy(backend Backend) *Proxy { return &Proxy{backend: backend} }
func (p *Proxy) Register(ctx context.Context, instance Instance) (Registration, error) {
	return p.backend.Register(ctx, instance)
}
func (p *Proxy) Watch(ctx context.Context, service string, update func([]Instance, error)) (func(), error) {
	return p.backend.Watch(ctx, service, update)
}
func (p *Proxy) List(ctx context.Context, service string) ([]Instance, error) {
	return p.backend.List(ctx, service)
}
func (p *Proxy) Healthy() bool { return p.backend.Healthy() }
func (p *Proxy) Close() error  { return p.backend.Close() }
