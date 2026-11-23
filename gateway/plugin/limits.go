package plugin

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"

	"golang.org/x/time/rate"
	"sitehub/gateway/internal/clientaddr"
)

type rateConfig struct {
	Rate           float64 `json:"rate"`
	Burst          int     `json:"burst"`
	IdleTTLSeconds int     `json:"idle_ttl_seconds"`
	MaxKeys        int     `json:"max_keys"`
}
type tokenEntry struct {
	limiter *rate.Limiter
	seen    time.Time
}
type tokenLimits struct {
	mu          sync.Mutex
	entries     map[string]*tokenEntry
	config      rateConfig
	nextCleanup time.Time
}

func remoteIP(r *http.Request) string {
	return clientaddr.From(r).Client
}

func rateLimit(raw json.RawMessage) (Middleware, error) {
	c := rateConfig{IdleTTLSeconds: 300, MaxKeys: 10000}
	if err := decode(raw, &c); err != nil {
		return nil, err
	}
	if c.Rate <= 0 || math.IsInf(c.Rate, 0) || math.IsNaN(c.Rate) || c.Rate > 1000000 || c.Burst < 1 || c.Burst > 1000000 || c.MaxKeys < 1 || c.MaxKeys > 1000000 || c.IdleTTLSeconds < 1 || c.IdleTTLSeconds > 86400 || float64(c.IdleTTLSeconds) < float64(c.Burst)/c.Rate {
		return nil, fmt.Errorf("positive bounded rate/burst/max_keys required; idle TTL must cover bucket refill")
	}
	state := &tokenLimits{entries: make(map[string]*tokenEntry), config: c}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !state.allow(remoteIP(r), time.Now()) {
				w.Header().Set("Retry-After", strconv.Itoa(int(math.Max(1, math.Ceil(1/c.Rate)))))
				respond(w, r, http.StatusTooManyRequests, "Request rate exceeded")
				return
			}
			next.ServeHTTP(w, r)
		})
	}, nil
}

func (s *tokenLimits) allow(key string, now time.Time) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !now.Before(s.nextCleanup) {
		for key, entry := range s.entries {
			if now.Sub(entry.seen) >= time.Duration(s.config.IdleTTLSeconds)*time.Second {
				delete(s.entries, key)
			}
		}
		s.nextCleanup = now.Add(time.Second)
	}
	entry, exists := s.entries[key]
	if !exists {
		// Reject excess new keys, rather than evicting an active bucket and
		// allowing clients to reset each other's counters under cardinality load.
		if len(s.entries) >= s.config.MaxKeys {
			return false
		}
		entry = &tokenEntry{limiter: rate.NewLimiter(rate.Limit(s.config.Rate), s.config.Burst)}
		s.entries[key] = entry
	}
	entry.seen = now
	return entry.limiter.AllowN(now, 1)
}

type countConfig struct {
	Count         int `json:"count"`
	WindowSeconds int `json:"window_seconds"`
	MaxKeys       int `json:"max_keys"`
}
type countEntry struct {
	used  int
	until time.Time
}
type countLimits struct {
	mu          sync.Mutex
	entries     map[string]countEntry
	config      countConfig
	nextCleanup time.Time
}

func limitCount(raw json.RawMessage) (Middleware, error) {
	c := countConfig{MaxKeys: 10000}
	if err := decode(raw, &c); err != nil {
		return nil, err
	}
	if c.Count < 1 || c.Count > 1000000 || c.WindowSeconds < 1 || c.WindowSeconds > 86400 || c.MaxKeys < 1 || c.MaxKeys > 1000000 {
		return nil, fmt.Errorf("positive bounded count/window_seconds/max_keys required")
	}
	state := &countLimits{entries: make(map[string]countEntry), config: c}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if allowed, retry := state.allow(remoteIP(r), time.Now()); !allowed {
				w.Header().Set("Retry-After", strconv.Itoa(retry))
				respond(w, r, http.StatusTooManyRequests, "Request quota exceeded")
				return
			}
			next.ServeHTTP(w, r)
		})
	}, nil
}

func (s *countLimits) allow(key string, now time.Time) (bool, int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !now.Before(s.nextCleanup) {
		for key, entry := range s.entries {
			if !now.Before(entry.until) {
				delete(s.entries, key)
			}
		}
		s.nextCleanup = now.Add(time.Second)
	}
	entry, exists := s.entries[key]
	if !exists || !now.Before(entry.until) {
		if !exists && len(s.entries) >= s.config.MaxKeys {
			return false, 1
		}
		entry = countEntry{until: now.Add(time.Duration(s.config.WindowSeconds) * time.Second)}
	}
	if entry.used >= s.config.Count {
		return false, int(math.Max(1, math.Ceil(entry.until.Sub(now).Seconds())))
	}
	entry.used++
	s.entries[key] = entry
	return true, 0
}
