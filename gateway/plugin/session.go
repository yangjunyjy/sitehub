package plugin

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
)

type browserSessionConfig struct {
	Audience         string   `json:"audience"`
	CookieName       string   `json:"cookie_name"`
	Origins          []string `json:"origins"`
	OriginCheckPaths []string `json:"origin_check_paths"`
}

func browserSession(raw json.RawMessage) (Middleware, error) {
	var config browserSessionConfig
	if err := decode(raw, &config); err != nil {
		return nil, err
	}
	if config.Audience != "" && config.Audience != "admin" && config.Audience != "site" {
		return nil, fmt.Errorf("audience must be admin or site when provided")
	}
	wanted := config.CookieName
	if wanted == "" {
		if config.Audience == "" {
			return nil, fmt.Errorf("cookie_name or audience is required")
		}
		wanted = "sitehub_" + config.Audience + "_session"
	}
	if err := (&http.Cookie{Name: wanted}).Valid(); err != nil {
		return nil, fmt.Errorf("cookie_name must be a valid HTTP cookie name")
	}
	if config.OriginCheckPaths == nil {
		// Preserve the existing browser login protection when upgrading older
		// configurations. An explicit empty list opts out of these path checks,
		// while cookie-authenticated write requests still require an origin.
		config.OriginCheckPaths = []string{"/v1/rbac/auth/login", "/v1/rbac/auth/register", "/v1/rbac/members/login", "/v1/rbac/members/register"}
	}
	originCheckPaths := make(map[string]bool, len(config.OriginCheckPaths))
	for _, value := range config.OriginCheckPaths {
		clean := path.Clean(value)
		if !strings.HasPrefix(value, "/") || strings.ContainsAny(value, "?#%\\* ") || strings.IndexFunc(value, func(c rune) bool { return c < 32 || c == 127 }) >= 0 || (clean != value && clean+"/" != value) || originCheckPaths[value] {
			return nil, fmt.Errorf("origin_check_paths must contain unique clean absolute paths without query strings, fragments or wildcards")
		}
		originCheckPaths[value] = true
	}
	if len(config.Origins) == 0 {
		return nil, fmt.Errorf("origins must not be empty")
	}
	origins := make(map[string]bool, len(config.Origins))
	for _, origin := range config.Origins {
		u, err := url.Parse(origin)
		if err != nil || u.Hostname() == "" || u.User != nil || u.Path != "" || u.RawQuery != "" || u.Fragment != "" || (u.Scheme != "http" && u.Scheme != "https") || origins[origin] {
			return nil, fmt.Errorf("origins must be unique HTTP origins without paths")
		}
		origins[origin] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if len(r.Header.Values("Authorization")) > 1 || len(r.Header.Values("Cookie")) > 1 || len(r.Header.Values("Origin")) > 1 {
				respond(w, r, http.StatusBadRequest, "Ambiguous credentials or origin")
				return
			}
			authorization, rawCookie := r.Header.Get("Authorization"), r.Header.Get("Cookie")
			token, found := "", false
			for _, item := range strings.Split(rawCookie, ";") {
				key, value, ok := strings.Cut(strings.TrimSpace(item), "=")
				if ok && key == wanted {
					if found {
						respond(w, r, http.StatusBadRequest, "Duplicate session cookie")
						return
					}
					token, found = value, true
				}
			}
			if len(token) > 4096 || strings.IndexFunc(token, func(c rune) bool { return c <= 32 || c >= 127 }) >= 0 || strings.ContainsAny(token, "\",\\") {
				respond(w, r, http.StatusBadRequest, "Invalid session cookie")
				return
			}
			unsafe := r.Method != http.MethodGet && r.Method != http.MethodHead && r.Method != http.MethodOptions
			login := originCheckPaths[r.URL.Path]
			origin := r.Header.Get("Origin")
			if unsafe && ((found && authorization == "") || (login && origin != "")) && !origins[origin] {
				respond(w, r, http.StatusForbidden, "Cross-origin submission rejected")
				return
			}
			r.Header.Del("Cookie")
			if authorization == "" && token != "" {
				r.Header.Set("Authorization", "Bearer "+token)
			}
			next.ServeHTTP(w, r)
		})
	}, nil
}

func stripCredentials(raw json.RawMessage) (Middleware, error) {
	if err := decode(raw, &struct{}{}); err != nil {
		return nil, err
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, name := range []string{"Cookie", "Authorization", "X-Service-Key", "X-Actor-ID", "X-User-ID", "X-Role", "X-Roles"} {
				r.Header.Del(name)
			}
			next.ServeHTTP(w, r)
		})
	}, nil
}

func requestID(raw json.RawMessage) (Middleware, error) {
	if err := decode(raw, &struct{}{}); err != nil {
		return nil, err
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := r.Header.Get("X-Request-ID")
			if len(r.Header.Values("X-Request-ID")) != 1 || len(id) == 0 || len(id) > 128 || strings.IndexFunc(id, func(c rune) bool {
				return !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-' || c == '_' || c == '.')
			}) >= 0 {
				var random [16]byte
				if _, err := rand.Read(random[:]); err != nil {
					respond(w, r, http.StatusServiceUnavailable, "Request identifier unavailable")
					return
				}
				id = hex.EncodeToString(random[:])
			}
			r.Header.Set("X-Request-ID", id)
			w.Header().Set("X-Request-ID", id)
			next.ServeHTTP(w, r)
		})
	}, nil
}
