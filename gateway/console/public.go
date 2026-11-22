package console

import (
	"net/http"
	"strings"
)

// WithoutSession keeps the private console credential out of public HTTP
// routes when the browser accesses both listeners on the same hostname.
func WithoutSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lines := r.Header.Values("Cookie")
		for i, line := range lines {
			if !strings.Contains(line, sessionCookie) {
				continue
			}
			fields := strings.Split(line, ";")
			kept := fields[:0]
			for _, field := range fields {
				name, _, _ := strings.Cut(strings.TrimSpace(field), "=")
				if strings.TrimSpace(name) != sessionCookie {
					kept = append(kept, field)
				}
			}
			lines[i] = strings.TrimSpace(strings.Join(kept, ";"))
		}
		if len(lines) == 1 && lines[0] == "" {
			r.Header.Del("Cookie")
		}
		next.ServeHTTP(w, r)
	})
}
