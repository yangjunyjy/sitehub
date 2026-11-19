// Package configcenter defines the configuration-center boundary.
package configcenter

import "context"

type Config struct {
	Enabled bool   `json:",default=false"`
	Key     string `json:",optional"`
}

// Backend uses logical keys relative to the configured namespace and group.
// Watch returns complete content, or an empty value when the key is deleted.
// Failed reads are delivered separately and must not replace valid content.
// Cancel is idempotent and may be invoked by the callback.
type Backend interface {
	Get(context.Context, string) ([]byte, error)
	Watch(context.Context, string, func([]byte, error)) (func(), error)
	Close() error
}

// Proxy delegates the configuration contract to the selected provider.
type Proxy struct{ backend Backend }

func NewProxy(backend Backend) *Proxy { return &Proxy{backend: backend} }
func (p *Proxy) Get(ctx context.Context, key string) ([]byte, error) {
	return p.backend.Get(ctx, key)
}
func (p *Proxy) Watch(ctx context.Context, key string, update func([]byte, error)) (func(), error) {
	return p.backend.Watch(ctx, key, update)
}
func (p *Proxy) Close() error { return p.backend.Close() }
