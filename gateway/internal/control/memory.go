package control

import (
	"bytes"
	"context"
	"sync"
)

// MemoryStore holds the applied snapshot in file mode. The configuration files
// remain authoritative; only an explicit, validated reload changes this store.
type MemoryStore struct {
	mu     sync.Mutex
	record Record
}

func (s *MemoryStore) Read(ctx context.Context) (Record, error) {
	if err := ctx.Err(); err != nil {
		return Record{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return Record{Revision: s.record.Revision, Data: bytes.Clone(s.record.Data)}, nil
}

func (s *MemoryStore) CompareAndSwap(ctx context.Context, expected int64, data []byte) (Record, error) {
	if err := ctx.Err(); err != nil {
		return Record{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if expected != s.record.Revision {
		return Record{}, ErrConflict
	}
	s.record = Record{Revision: expected + 1, Data: bytes.Clone(data)}
	return Record{Revision: s.record.Revision, Data: bytes.Clone(s.record.Data)}, nil
}
