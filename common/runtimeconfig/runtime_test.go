package runtimeconfig

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sync"
	"testing"
)

func TestReloadRetainsInflightAndRejectsInvalidCandidate(t *testing.T) {
	m, err := New(Config{MaxConcurrentRequests: 2}, "v1", "test")
	if err != nil {
		t.Fatal(err)
	}
	first, _ := m.Enter()
	defer first()
	if err = m.Commit(Config{MaxConcurrentRequests: 1}, "v2"); err != nil {
		t.Fatal(err)
	}
	if _, err = m.Enter(); status.Code(err) != codes.ResourceExhausted {
		t.Fatalf("existing request was not counted: %v", err)
	}
	if err = m.Commit(Config{MaxConcurrentRequests: -1}, "invalid"); err == nil {
		t.Fatal("invalid snapshot accepted")
	}
	if m.Status().AppliedVersion != "v2" {
		t.Fatal("bad update replaced working version")
	}
	if err = m.Commit(Config{Maintenance: true}, "v3"); err != nil {
		t.Fatal(err)
	}
	if _, err = m.Enter(); status.Code(err) != codes.Unavailable {
		t.Fatal(err)
	}
}
func TestSnapshotAtomicUnderConcurrentTraffic(t *testing.T) {
	m, _ := New(Config{}, "v1", "test")
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for n := 0; n < 1000; n++ {
				if release, err := m.Enter(); err == nil {
					release()
				}
				_ = m.Status()
			}
		}()
	}
	for n := 0; n < 100; n++ {
		_ = m.Commit(Config{MaxConcurrentRequests: int64(n % 5)}, "next")
	}
	wg.Wait()
	if m.active.Load() != 0 {
		t.Fatal("request permits leaked")
	}
}
