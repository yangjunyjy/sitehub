package plugin

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"

	"golang.org/x/net/http/httpguts"
)

// proxyRewrite changes an upstream request after routing and identity checks.
// The middleware owns a cloned request; it never changes the original request,
// the destination upstream, or the URL query string.
func proxyRewrite(raw json.RawMessage) (Middleware, error) {
	var config struct {
		StripPrefix string            `json:"strip_prefix"`
		Host        string            `json:"host"`
		Headers     map[string]string `json:"headers"`
	}
	if err := decode(raw, &config); err != nil {
		return nil, err
	}
	if prefix := config.StripPrefix; prefix != "" && (!strings.HasPrefix(prefix, "/") || strings.ContainsAny(prefix, "?#%\\\x00\r\n") || path.Clean(prefix) != prefix) {
		return nil, fmt.Errorf("strip_prefix must be a clean absolute path without a trailing slash")
	}
	if config.Host != "" && !validRewriteHost(config.Host) {
		return nil, fmt.Errorf("host must be a hostname or IP address with an optional port")
	}
	headers := make(http.Header, len(config.Headers))
	for name, value := range config.Headers {
		if !httpguts.ValidHeaderFieldName(name) || !httpguts.ValidHeaderFieldValue(value) {
			return nil, fmt.Errorf("invalid request header")
		}
		canonical := http.CanonicalHeaderKey(name)
		if _, exists := headers[canonical]; exists {
			return nil, fmt.Errorf("duplicate request header %s", canonical)
		}
		lower := strings.ToLower(name)
		switch lower {
		case "host", "content-length", "connection", "transfer-encoding", "upgrade", "te", "trailer", "keep-alive", "proxy-connection", "proxy-authenticate", "proxy-authorization",
			"authorization", "cookie", "x-service-key", "x-actor-id", "x-user-id", "x-role", "x-roles", "forwarded", "x-real-ip", "x-original-url", "x-rewrite-url", "x-request-id":
			return nil, fmt.Errorf("unsupported request header %s", name)
		}
		if strings.HasPrefix(lower, "x-forwarded-") {
			return nil, fmt.Errorf("unsupported request header %s", name)
		}
		headers.Set(canonical, value)
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rewritten := r.Clone(r.Context())
			prefix := config.StripPrefix
			if prefix != "" && prefix != "/" && (r.URL.Path == prefix || strings.HasPrefix(r.URL.Path, prefix+"/")) {
				// Slice the original escaped path at the decoded prefix boundary.
				// Re-encoding Path would change %xx spelling and encoded characters
				// that are deliberately preserved by the reverse proxy.
				escaped := r.URL.EscapedPath()
				offset := escapedPrefixBytes(escaped, len(prefix))
				rewritten.URL.Path = r.URL.Path[len(prefix):]
				rewritten.URL.RawPath = escaped[offset:]
				if rewritten.URL.Path == "" {
					rewritten.URL.Path, rewritten.URL.RawPath = "/", "/"
				}
			}
			if config.Host != "" {
				rewritten.Host = config.Host
			}
			for name, values := range headers {
				rewritten.Header.Set(name, values[0])
			}
			next.ServeHTTP(w, rewritten)
		})
	}, nil
}

func escapedPrefixBytes(escaped string, decodedLength int) int {
	offset := 0
	for decoded := 0; decoded < decodedLength; decoded++ {
		if escaped[offset] == '%' {
			offset += 3
		} else {
			offset++
		}
	}
	return offset
}

func validRewriteHost(value string) bool {
	if strings.ContainsAny(value, "/?#@\\%\x00\r\n\t ") {
		return false
	}
	u, err := url.Parse("http://" + value)
	if err != nil || u.Host != value || u.Hostname() == "" || u.User != nil || strings.HasSuffix(value, ":") {
		return false
	}
	if port := u.Port(); port != "" {
		n, err := strconv.Atoi(port)
		if err != nil || n < 1 || n > 65535 {
			return false
		}
	}
	host := u.Hostname()
	if ip := net.ParseIP(host); ip != nil {
		return !strings.Contains(host, ":") || strings.HasPrefix(value, "[")
	}
	if len(host) > 253 || strings.HasPrefix(value, "[") {
		return false
	}
	for _, label := range strings.Split(strings.TrimSuffix(host, "."), ".") {
		if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for _, ch := range label {
			if !(ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' || ch >= '0' && ch <= '9' || ch == '-' || ch == '_') {
				return false
			}
		}
	}
	return true
}
