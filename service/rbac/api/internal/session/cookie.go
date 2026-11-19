// Package session adapts successful RBAC API sessions to browser cookies.
// It never validates a session or replaces RPC authorization.
package session

import (
	"context"
	"net/http"
	"time"
)

type Config struct {
	Secure bool `json:",default=true"`
}
type key struct{}
type sink func(audience, token string, expires int64)

func Middleware(c Config) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			write := sink(func(audience, token string, expires int64) {
				if audience != "admin" && audience != "site" {
					return
				}
				age := -1
				if token != "" {
					age = int(min(int64(86400), expires-time.Now().Unix()))
					if age <= 0 {
						token = ""
						age = -1
					}
				}
				w.Header().Set("Cache-Control", "no-store")
				http.SetCookie(w, &http.Cookie{
					Name:     "sitehub_" + audience + "_session",
					Value:    token,
					Path:     "/",
					HttpOnly: true,
					Secure:   c.Secure,
					SameSite: http.SameSiteLaxMode,
					MaxAge:   age})
			})
			next(w, r.WithContext(context.WithValue(r.Context(), key{}, write)))
		}
	}
}

// Issue is called only after the RPC and response conversion have succeeded.
func Issue(ctx context.Context, audience, token string, expires int64) {
	if write, ok := ctx.Value(key{}).(sink); ok {
		write(audience, token, expires)
	}
}
func Clear(ctx context.Context, audience string) {
	Issue(ctx, audience, "", 0)
}
