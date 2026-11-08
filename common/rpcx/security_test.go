package rpcx

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestHTTPIdentityCannotSupplyServiceCredentials(t *testing.T) {
	request := httptest.NewRequest("GET", "/", nil)
	request.Header.Set("Authorization", "Bearer user-credential")
	request.Header.Set("X-Service-Key", "forged")
	request.Header.Set("X-Actor-ID", "1")
	HTTPContext(func(_ http.ResponseWriter, r *http.Request) {
		md, _ := metadata.FromOutgoingContext(Outgoing(r.Context(), "configured-service-key-with-32-characters"))
		if md.Get("authorization")[0] != "user-credential" || md.Get("x-service-key")[0] == "forged" || len(md.Get("x-actor-id")) != 0 {
			t.Fatal("untrusted identity crossed the service boundary")
		}
	})(httptest.NewRecorder(), request)
}
func TestInternalAuthenticationFailsClosed(t *testing.T) {
	key := "configured-service-key-with-32-characters"
	for _, supplied := range []string{"", "forged"} {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-service-key", supplied))
		if status.Code(RequireService(ctx, key)) != codes.Unauthenticated {
			t.Fatal("invalid key accepted")
		}
	}
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-service-key", key))
	if err := RequireService(ctx, key); err != nil {
		t.Fatal(err)
	}
}

func TestClientIPTrustBoundary(t *testing.T) {
	for _, tt := range []struct {
		name, remote, forwarded, want string
		trusted                       []string
	}{
		{"direct spoof", "198.51.100.5:1234", "203.0.113.7", "198.51.100.5", nil},
		{"trusted gateway", "10.0.0.2:1234", "203.0.113.7", "203.0.113.7", []string{"10.0.0.0/24"}},
		{"outside trusted subnet", "10.0.1.2:1234", "203.0.113.7", "10.0.1.2", []string{"10.0.0.0/24"}},
		{"invalid forwarded value", "10.0.0.2:1234", "127.0.0.1, 203.0.113.7", "10.0.0.2", []string{"10.0.0.0/24"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/", nil)
			r.RemoteAddr = tt.remote
			r.Header.Set("X-Real-IP", tt.forwarded)
			HTTPContextWithProxies(tt.trusted)(func(_ http.ResponseWriter, r *http.Request) {
				md, _ := metadata.FromOutgoingContext(Outgoing(r.Context(), "key"))
				if got := md.Get("x-client-ip"); len(got) != 1 || got[0] != tt.want {
					t.Fatalf("client IP = %v, want %s", got, tt.want)
				}
			})(httptest.NewRecorder(), r)
		})
	}
}
