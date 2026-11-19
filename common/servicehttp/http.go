package servicehttp

import (
	"bytes"
	"context"
	"crypto/subtle"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sitehub/common/rpcx"
)

var transport = &http.Client{Timeout: 5 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}

func Authorized(request *http.Request, key string) bool {
	return len(key) >= 32 && subtle.ConstantTimeCompare([]byte(request.Header.Get("X-Service-Key")), []byte(key)) == 1
}

func Write(w http.ResponseWriter, value any, err error) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		failure := status.Convert(rpcx.Error(err))
		code := http.StatusInternalServerError
		switch failure.Code() {
		case codes.InvalidArgument:
			code = http.StatusBadRequest
		case codes.NotFound:
			code = http.StatusNotFound
		case codes.PermissionDenied:
			code = http.StatusForbidden
		case codes.Unauthenticated:
			code = http.StatusUnauthorized
		case codes.Aborted, codes.FailedPrecondition, codes.AlreadyExists:
			code = http.StatusConflict
		case codes.Unavailable:
			code = http.StatusServiceUnavailable
		}
		message := http.StatusText(code)
		if code < 500 {
			message = failure.Message()
		}
		w.WriteHeader(code)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": message})
		return
	}
	if value == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	_ = json.NewEncoder(w).Encode(value)
}

func Do(ctx context.Context, base, key, method, path string, in, out any) error {
	address, err := url.Parse(base)
	if err != nil || (address.Scheme != "http" && address.Scheme != "https") || address.Host == "" || address.User != nil || address.RawQuery != "" || address.Fragment != "" {
		return status.Error(codes.Unavailable, "invalid internal service URL")
	}
	var body io.Reader
	if in != nil {
		data, err := json.Marshal(in)
		if err != nil {
			return err
		}
		body = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, strings.TrimRight(base, "/")+path, body)
	if err != nil {
		return err
	}
	req.Header.Set("X-Service-Key", key)
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	rpcx.ForwardIdentity(ctx, req.Header)
	response, err := transport.Do(req)
	if err != nil {
		return status.Error(codes.Unavailable, "internal service unavailable")
	}
	defer response.Body.Close()
	decoder := json.NewDecoder(io.LimitReader(response.Body, 2<<20))
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		var failure struct{ Message string }
		_ = decoder.Decode(&failure)
		code := codes.Unavailable
		switch response.StatusCode {
		case 400:
			code = codes.InvalidArgument
		case 401:
			code = codes.Unauthenticated
		case 403:
			code = codes.PermissionDenied
		case 404:
			code = codes.NotFound
		case 409:
			code = codes.Aborted
		}
		return status.Error(code, failure.Message)
	}
	if out != nil {
		return decoder.Decode(out)
	}
	return nil
}
