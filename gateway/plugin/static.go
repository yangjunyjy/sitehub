package plugin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/net/http/httpguts"
)

func staticResponse(raw json.RawMessage) (Middleware, error) {
	var config struct {
		Status  int               `json:"status"`
		Body    string            `json:"body"`
		Headers map[string]string `json:"headers"`
	}
	if err := decode(raw, &config); err != nil {
		return nil, err
	}
	if config.Status < 200 || config.Status > 599 || len(config.Body) > 64<<10 {
		return nil, fmt.Errorf("status must be 200..599 and body at most 64 KiB")
	}
	if (config.Status == 204 || config.Status == 304) && config.Body != "" {
		return nil, fmt.Errorf("204 and 304 responses must not have a body")
	}
	for name, value := range config.Headers {
		if !httpguts.ValidHeaderFieldName(name) || !httpguts.ValidHeaderFieldValue(value) {
			return nil, fmt.Errorf("invalid response header")
		}
		switch strings.ToLower(name) {
		case "content-length", "connection", "transfer-encoding", "upgrade", "trailer", "keep-alive", "proxy-authenticate", "proxy-authorization":
			return nil, fmt.Errorf("unsupported response header %s", name)
		}
	}
	return func(_ http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for name, value := range config.Headers {
				w.Header().Set(name, value)
			}
			if config.Status != 204 && config.Status != 304 {
				w.Header().Set("Content-Length", strconv.Itoa(len(config.Body)))
			}
			w.WriteHeader(config.Status)
			if r.Method != http.MethodHead && config.Body != "" {
				_, _ = w.Write([]byte(config.Body))
			}
		})
	}, nil
}
