package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"math"
	"net"
	"sort"
	"strings"
	"sync"

	"github.com/google/uuid"
	clientv3 "go.etcd.io/etcd/client/v3"
	"sitehub/common/registry"
)

type Registry struct{ *base }

func NewRegistry(options Options) (*Registry, error) {
	b, err := connect(options)
	if err != nil {
		return nil, err
	}
	return &Registry{base: b}, nil
}

func (r *Registry) List(ctx context.Context, service string) ([]registry.Instance, error) {
	if !validKey(service) {
		return nil, fmt.Errorf("invalid discovery service name")
	}
	if err := r.begin(); err != nil {
		return nil, err
	}
	defer r.wg.Done()
	response, err := r.get(ctx, r.options.Prefix+service+"/", true)
	if err != nil {
		return nil, err
	}
	return instances(snapshot(response), service), nil
}

func (r *Registry) Watch(ctx context.Context, service string, update func([]registry.Instance, error)) (func(), error) {
	if !validKey(service) || update == nil {
		return nil, fmt.Errorf("discovery service name and callback are required")
	}
	return r.watch(ctx, r.options.Prefix+service+"/", true, func(values map[string][]byte, err error) { update(instances(values, service), err) })
}

func instances(values map[string][]byte, service string) []registry.Instance {
	result := make([]registry.Instance, 0, len(values))
	for _, value := range values {
		var instance registry.Instance
		if json.Unmarshal(value, &instance) == nil && instance.Service == service && validInstance(instance) {
			result = append(result, instance)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

func validInstance(instance registry.Instance) bool {
	ip := net.ParseIP(instance.IP)
	return validKey(instance.Service) && validKey(instance.ID) && ip != nil && !ip.IsUnspecified() && !ip.IsMulticast() && instance.Port > 0 && instance.Port <= 65535 && !math.IsNaN(instance.Weight) && !math.IsInf(instance.Weight, 0) && instance.Weight >= 0
}

func (r *Registry) Register(parent context.Context, instance registry.Instance) (registry.Registration, error) {
	if instance.ID == "" {
		instance.ID = uuid.NewString()
	}
	if !validInstance(instance) {
		return nil, fmt.Errorf("registration requires a service, ID, IP, valid port and finite nonnegative weight")
	}
	instance.Metadata = maps.Clone(instance.Metadata)
	if instance.Metadata == nil {
		instance.Metadata = make(map[string]string)
	}
	if _, exists := instance.Metadata["protocol"]; !exists {
		instance.Metadata["protocol"] = "http"
		if strings.HasSuffix(instance.Service, "-rpc") {
			instance.Metadata["protocol"] = "grpc"
		}
	}
	value, err := json.Marshal(instance)
	if err != nil {
		return nil, fmt.Errorf("encode registration failed")
	}
	if err = r.begin(); err != nil {
		return nil, err
	}
	ctx, cancel := r.lifetime(parent)
	registration := &registration{base: r.base, ctx: ctx, cancel: cancel, key: r.options.Prefix + instance.Service + "/" + instance.ID, value: string(value), done: make(chan struct{})}
	lease, updates, err := registration.establish()
	if err != nil {
		cancel()
		r.wg.Done()
		return nil, err
	}
	go registration.run(lease, updates)
	return registration, nil
}

func (r *Registry) Healthy() bool { return r.healthy.Load() && r.ctx.Err() == nil }
func (r *Registry) Close() error  { return r.close() }

type registration struct {
	*base
	ctx        context.Context
	cancel     func()
	key, value string
	done       chan struct{}
	once       sync.Once
	closeErr   error
}

func (r *registration) establish() (clientv3.LeaseID, <-chan *clientv3.LeaseKeepAliveResponse, error) {
	ctx, cancel := r.operation(r.ctx)
	defer cancel()
	lease, err := r.client.Grant(ctx, r.options.LeaseTTL)
	if err != nil {
		return 0, nil, fmt.Errorf("grant registration lease failed")
	}
	if _, err = r.client.Put(ctx, r.key, r.value, clientv3.WithLease(lease.ID)); err != nil {
		_ = r.revoke(lease.ID)
		return 0, nil, fmt.Errorf("publish registration failed")
	}
	updates, err := r.client.KeepAlive(r.ctx, lease.ID)
	if err != nil || updates == nil {
		_ = r.revoke(lease.ID)
		return 0, nil, fmt.Errorf("keep registration lease alive failed")
	}
	r.healthy.Store(true)
	return lease.ID, updates, nil
}

func (r *registration) revoke(lease clientv3.LeaseID) error {
	ctx, cancel := context.WithTimeout(context.Background(), r.options.RequestTimeout)
	defer cancel()
	if _, err := r.client.Revoke(ctx, lease); err != nil {
		return fmt.Errorf("revoke registration lease failed")
	}
	return nil
}

func (r *registration) run(lease clientv3.LeaseID, updates <-chan *clientv3.LeaseKeepAliveResponse) {
	defer r.wg.Done()
	defer close(r.done)
	defer r.cancel()
	defer func() { r.closeErr = r.revoke(lease) }()
	for {
		select {
		case <-r.ctx.Done():
			return
		case response, ok := <-updates:
			if ok && response != nil && response.TTL > 0 {
				r.healthy.Store(true)
				continue
			}
			r.healthy.Store(false)
			_ = r.revoke(lease)
			for retry(r.ctx) {
				newLease, newUpdates, err := r.establish()
				if err != nil {
					continue
				}
				lease, updates = newLease, newUpdates
				break
			}
			if r.ctx.Err() != nil {
				return
			}
		}
	}
}

func (r *registration) Close() error {
	r.once.Do(r.cancel)
	<-r.done
	return r.closeErr
}
