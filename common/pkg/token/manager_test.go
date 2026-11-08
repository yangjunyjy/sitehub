package token

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
)

func TestOnceTokenHasOnlyOneConcurrentConsumer(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	manager := NewRedisTokenManager(client, nil)
	issued, err := manager.Generate(context.Background(), 42, "member", nil)
	if err != nil {
		t.Fatal(err)
	}
	var accepted atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			value, err := manager.Verify(context.Background(), issued.TokenID)
			if err == nil {
				accepted.Add(1)
				if value.UserID != 42 {
					t.Error("wrong owner")
				}
			} else if !errors.Is(err, ErrTokenNotFound) {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	if got := accepted.Load(); got != 1 {
		t.Fatalf("accepted %d consumers", got)
	}
}
