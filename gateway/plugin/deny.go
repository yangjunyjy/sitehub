package plugin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strings"
)

// denyPaths applies an explicitly configured path policy after route selection.
// The engine checks request-path ambiguity before matching and running plugins.
func denyPaths(raw json.RawMessage) (Middleware, error) {
	var config struct {
		Paths  []string `json:"paths"`
		Status *int     `json:"status"`
	}
	if err := decode(raw, &config); err != nil {
		return nil, err
	}
	if len(config.Paths) == 0 || len(config.Paths) > 64 {
		return nil, fmt.Errorf("paths must contain 1..64 exact paths or trailing /* patterns")
	}
	status := http.StatusNotFound
	if config.Status != nil {
		status = *config.Status
	}
	if status < 400 || status > 599 {
		return nil, fmt.Errorf("status must be 400..599")
	}
	seen := make(map[string]bool, len(config.Paths))
	for _, pattern := range config.Paths {
		if !validDeniedPath(pattern) || seen[pattern] {
			return nil, fmt.Errorf("paths must be unique clean absolute paths with an optional trailing /*")
		}
		seen[pattern] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, pattern := range config.Paths {
				if r.URL.Path == pattern || strings.HasSuffix(pattern, "/*") && strings.HasPrefix(r.URL.Path, strings.TrimSuffix(pattern, "*")) {
					message := "Request rejected"
					if status == http.StatusNotFound {
						message = "Not found"
					}
					respond(w, r, status, message)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}, nil
}

func validDeniedPath(value string) bool {
	if value == "" || !strings.HasPrefix(value, "/") || strings.ContainsAny(value, "?#%\\") || strings.IndexFunc(value, func(ch rune) bool { return ch < 32 || ch == 127 }) >= 0 {
		return false
	}
	if strings.HasSuffix(value, "/*") {
		value = strings.TrimSuffix(value, "*")
	}
	if strings.Contains(value, "*") || strings.Contains(value, "//") {
		return false
	}
	return path.Clean(value) == value || strings.HasSuffix(value, "/") && path.Clean(value)+"/" == value
}
