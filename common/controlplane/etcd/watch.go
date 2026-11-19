package etcd

import (
	"bytes"
	"context"
	"fmt"
	"maps"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// watch starts from the revision of the first read, so a write between the
// initial snapshot and Watch creation remains observable. Reconnects always
// resnapshot, which also recovers from compacted revisions.
func (b *base) watch(parent context.Context, key string, prefix bool, update func(map[string][]byte, error)) (func(), error) {
	if update == nil {
		return nil, fmt.Errorf("watch callback is required")
	}
	if err := b.begin(); err != nil {
		return nil, err
	}
	ctx, cancel := b.lifetime(parent)
	initial, err := b.get(ctx, key, prefix)
	if err != nil {
		cancel()
		b.wg.Done()
		return nil, err
	}
	go func() {
		defer b.wg.Done()
		defer cancel()
		current := snapshot(initial)
		revision := initial.Header.Revision
		if ctx.Err() != nil {
			return
		}
		update(cloneValues(current), nil)
		for ctx.Err() == nil {
			watchCtx, stopWatch := context.WithCancel(clientv3.WithRequireLeader(ctx))
			opts := []clientv3.OpOption{clientv3.WithRev(revision + 1), clientv3.WithProgressNotify()}
			if prefix {
				opts = append(opts, clientv3.WithPrefix())
			}
			changes := b.client.Watch(watchCtx, key, opts...)
			// A bounded linearizable read detects a silent broken stream and
			// keeps the health/freshness contract independent of server push.
			tick := time.NewTicker(time.Second)
			failed := false
			for !failed && ctx.Err() == nil {
				select {
				case <-ctx.Done():
				case response, ok := <-changes:
					if !ok || response.Err() != nil || response.Canceled {
						failed = true
						break
					}
					changed := false
					nextRevision := revision
					for _, event := range response.Events {
						if event.Kv.ModRevision <= revision {
							continue
						}
						name := string(event.Kv.Key)
						nextRevision = max(nextRevision, event.Kv.ModRevision)
						if event.Type == clientv3.EventTypeDelete {
							if _, exists := current[name]; exists {
								delete(current, name)
								changed = true
							}
						} else if previous, exists := current[name]; !exists || !bytes.Equal(previous, event.Kv.Value) {
							current[name] = bytes.Clone(event.Kv.Value)
							changed = true
						}
					}
					// Historical event batches may carry a newer server header.
					// Advance only through events delivered, or a progress barrier.
					if response.IsProgressNotify() {
						nextRevision = max(nextRevision, response.Header.Revision)
					}
					revision = nextRevision
					if changed && ctx.Err() == nil {
						update(cloneValues(current), nil)
					}
				case <-tick.C:
					response, err := b.get(ctx, key, prefix)
					if err != nil {
						failed = true
						break
					}
					fresh := snapshot(response)
					changed := !maps.EqualFunc(current, fresh, bytes.Equal)
					current, revision = fresh, response.Header.Revision
					if changed && ctx.Err() == nil {
						update(cloneValues(current), nil)
					}
				}
			}
			tick.Stop()
			stopWatch()
			if ctx.Err() != nil {
				return
			}
			b.healthy.Store(false)
			update(nil, fmt.Errorf("control-plane watch interrupted"))
			for retry(ctx) {
				response, err := b.get(ctx, key, prefix)
				if err != nil {
					continue
				}
				current, revision = snapshot(response), response.Header.Revision
				if ctx.Err() == nil {
					update(cloneValues(current), nil)
				}
				break
			}
		}
	}()
	return cancel, nil
}

func snapshot(response *clientv3.GetResponse) map[string][]byte {
	values := make(map[string][]byte, len(response.Kvs))
	for _, pair := range response.Kvs {
		values[string(pair.Key)] = bytes.Clone(pair.Value)
	}
	return values
}

func cloneValues(values map[string][]byte) map[string][]byte {
	result := make(map[string][]byte, len(values))
	for key, value := range values {
		result[key] = bytes.Clone(value)
	}
	return result
}
