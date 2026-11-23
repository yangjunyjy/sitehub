package control

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

var ErrConflict = errors.New("configuration revision changed; fetch again before publishing")

type Record struct {
	Revision int64
	Data     []byte
}

// A release receipt shares the configuration transaction's revision. Its value
// binds the submitted manifest and base revision to the publisher's release ID.
type receiptStore interface {
	PublishRelease(context.Context, int64, []byte, string) (Record, error)
	Receipt(context.Context, string) (Record, error)
}
type Store interface {
	Read(context.Context) (Record, error)
	CompareAndSwap(context.Context, int64, []byte) (Record, error)
}
type EtcdStore struct {
	Client *clientv3.Client
	Key    string
}

func (s *EtcdStore) Read(ctx context.Context) (Record, error) {
	record, _, err := s.read(ctx, s.Key)
	return record, err
}
func (s *EtcdStore) read(ctx context.Context, key string) (Record, int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	resp, err := s.Client.Get(ctx, key)
	if err != nil {
		return Record{}, 0, fmt.Errorf("gateway configuration store unavailable")
	}
	if len(resp.Kvs) == 0 {
		return Record{}, resp.Header.Revision, nil
	}
	return Record{Revision: resp.Kvs[0].ModRevision, Data: resp.Kvs[0].Value}, resp.Header.Revision, nil
}
func (s *EtcdStore) CompareAndSwap(ctx context.Context, expected int64, data []byte) (Record, error) {
	return s.PublishRelease(ctx, expected, data, "")
}
func (s *EtcdStore) Receipt(ctx context.Context, generation string) (Record, error) {
	record, _, err := s.read(ctx, s.Key+"/releases/"+generation)
	return record, err
}
func (s *EtcdStore) PublishRelease(ctx context.Context, expected int64, data []byte, generation string) (Record, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	conditions := []clientv3.Cmp{clientv3.Compare(clientv3.ModRevision(s.Key), "=", expected)}
	updates := []clientv3.Op{clientv3.OpPut(s.Key, string(data))}
	if generation != "" {
		key := s.Key + "/releases/" + generation
		conditions = append(conditions, clientv3.Compare(clientv3.Version(key), "=", 0))
		updates = append(updates, clientv3.OpPut(key, fmt.Sprintf("%d:%x", expected, sha256.Sum256(data))))
	}
	resp, err := s.Client.Txn(ctx).If(conditions...).Then(updates...).Commit()
	if err != nil {
		return Record{}, fmt.Errorf("configuration commit could not be confirmed; read current revision before retrying")
	}
	if !resp.Succeeded {
		return Record{}, ErrConflict
	}
	return Record{Revision: resp.Header.Revision, Data: append([]byte(nil), data...)}, nil
}

// Watch always resynchronizes with a linearizable read after a disconnect or
// compaction. Starting at read revision+1 prevents the read/watch race.
func (s *EtcdStore) Watch(ctx context.Context, apply func(Record), failure func(error)) {
	for ctx.Err() == nil {
		record, readRevision, err := s.read(ctx, s.Key)
		if err == nil {
			apply(record)
			stream := s.Client.Watch(ctx, s.Key, clientv3.WithRev(readRevision+1))
			for response := range stream {
				if response.Err() != nil {
					err = response.Err()
					break
				}
				for _, event := range response.Events {
					if event.Type == clientv3.EventTypeDelete {
						failure(fmt.Errorf("active gateway configuration was deleted; retaining applied snapshot"))
						continue
					}
					apply(Record{Revision: event.Kv.ModRevision, Data: append([]byte(nil), event.Kv.Value...)})
				}
			}
		}
		if ctx.Err() != nil {
			return
		}
		failure(fmt.Errorf("configuration watch disconnected; retaining applied snapshot"))
		timer := time.NewTimer(time.Second)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
	}
}
