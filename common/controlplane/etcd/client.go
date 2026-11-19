// Package etcd implements the control-plane contracts. Provider wire types are
// confined to this adapter and never appear in application-facing interfaces.
package etcd

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

type Options struct {
	Endpoints                   []string
	Username, Password          string
	Prefix                      string
	DialTimeout, RequestTimeout time.Duration
	LeaseTTL                    int64
}

// client is deliberately small enough to exercise lease and watch failures
// deterministically without running an external service.
type client interface {
	Get(context.Context, string, ...clientv3.OpOption) (*clientv3.GetResponse, error)
	Put(context.Context, string, string, ...clientv3.OpOption) (*clientv3.PutResponse, error)
	Watch(context.Context, string, ...clientv3.OpOption) clientv3.WatchChan
	Grant(context.Context, int64) (*clientv3.LeaseGrantResponse, error)
	Revoke(context.Context, clientv3.LeaseID) (*clientv3.LeaseRevokeResponse, error)
	KeepAlive(context.Context, clientv3.LeaseID) (<-chan *clientv3.LeaseKeepAliveResponse, error)
	Close() error
}

type base struct {
	client   client
	options  Options
	ctx      context.Context
	cancel   context.CancelFunc
	healthy  atomic.Bool
	mu       sync.Mutex
	closed   bool
	wg       sync.WaitGroup
	once     sync.Once
	closeErr error
}

func connect(options Options) (*base, error) {
	c, err := clientv3.New(clientv3.Config{
		Endpoints: options.Endpoints, Username: options.Username, Password: options.Password,
		DialTimeout: options.DialTimeout, Logger: zap.NewNop(),
	})
	if err != nil {
		return nil, fmt.Errorf("create control-plane client failed")
	}
	return newBase(c, options), nil
}

func newBase(c client, options Options) *base {
	ctx, cancel := context.WithCancel(context.Background())
	b := &base{client: c, options: options, ctx: ctx, cancel: cancel}
	return b
}

func (b *base) begin() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return fmt.Errorf("control-plane client is closed")
	}
	b.wg.Add(1)
	return nil
}

func (b *base) operation(parent context.Context) (context.Context, func()) {
	ctx, cancel := context.WithTimeout(parent, b.options.RequestTimeout)
	stop := context.AfterFunc(b.ctx, cancel)
	return ctx, func() { stop(); cancel() }
}

func (b *base) lifetime(parent context.Context) (context.Context, func()) {
	ctx, cancel := context.WithCancel(parent)
	stop := context.AfterFunc(b.ctx, cancel)
	return ctx, func() { stop(); cancel() }
}

func (b *base) get(ctx context.Context, key string, prefix bool) (*clientv3.GetResponse, error) {
	request, cancel := b.operation(ctx)
	defer cancel()
	var opts []clientv3.OpOption
	if prefix {
		opts = append(opts, clientv3.WithPrefix())
	}
	response, err := b.client.Get(request, key, opts...)
	if err != nil {
		b.healthy.Store(false)
		if request.Err() != nil {
			return nil, fmt.Errorf("control-plane read failed: %w", request.Err())
		}
		return nil, fmt.Errorf("control-plane read failed")
	}
	b.healthy.Store(true)
	return response, nil
}

func (b *base) close() error {
	b.once.Do(func() {
		b.mu.Lock()
		b.closed = true
		b.cancel()
		b.mu.Unlock()
		b.wg.Wait()
		b.closeErr = b.client.Close()
	})
	return b.closeErr
}

func retry(ctx context.Context) bool {
	timer := time.NewTimer(250 * time.Millisecond)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
