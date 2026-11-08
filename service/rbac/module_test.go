package rbac

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"

	"sitehub/common/config"
)

func TestTokenLifetimeComesFromModuleConfiguration(t *testing.T) {
	server := miniredis.RunT(t)
	cache := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = cache.Close() })
	cfg := &config.Config{}
	cfg.OnceToken.ExpireTime = 42
	module := New(nil, cache, cfg, nil)
	issued, err := module.Tokens.Generate(context.Background(), 1, "test-user", nil)
	if err != nil {
		t.Fatal(err)
	}
	server.FastForward(43 * time.Second)
	if _, err := module.Tokens.Verify(context.Background(), issued.TokenID); err == nil {
		t.Fatal("configured token expiry was ignored")
	}
}
