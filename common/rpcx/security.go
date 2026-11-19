package rpcx

import (
	"context"
	"crypto/subtle"
	"net"
	"net/http"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type identityKey struct{}
type headers struct{ Token, TOTP, IP, Agent string }

// HTTPContext accepts only end-user credentials, never an internal service key.
func HTTPContext(next http.HandlerFunc) http.HandlerFunc {
	return HTTPContextWithProxies(nil)(next)
}

// HTTPContextWithProxies accepts X-Real-IP only from explicitly configured proxies.
// The gateway must replace this header with the connection's actual client address.
func HTTPContextWithProxies(cidrs []string) func(http.HandlerFunc) http.HandlerFunc {
	trusted := make([]*net.IPNet, 0, len(cidrs))
	for _, cidr := range cidrs {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			panic(err)
		}
		trusted = append(trusted, network)
	}
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			authorization := r.Header.Get("Authorization")
			token := ""
			if strings.HasPrefix(authorization, "Bearer ") {
				token = strings.TrimSpace(strings.TrimPrefix(authorization, "Bearer "))
			}
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				ip = r.RemoteAddr
			}
			for _, network := range trusted {
				if network.Contains(net.ParseIP(ip)) {
					if forwarded := net.ParseIP(r.Header.Get("X-Real-IP")); forwarded != nil {
						ip = forwarded.String()
					}
					break
				}
			}
			h := headers{token, r.Header.Get("X-TOTP-Token"), ip, r.UserAgent()}
			next(w, r.WithContext(context.WithValue(r.Context(), identityKey{}, h)))
		}
	}
}

func HTTPValue(ctx context.Context, name string) string {
	h, _ := ctx.Value(identityKey{}).(headers)
	switch name {
	case "authorization":
		return h.Token
	case "x-totp-token":
		return h.TOTP
	case "x-client-ip":
		return h.IP
	case "x-client-agent":
		return h.Agent
	}
	return ""
}

func Incoming(ctx context.Context, name string) string {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if values := md.Get(name); len(values) > 0 {
			return values[0]
		}
	}
	return ""
}

func Outgoing(ctx context.Context, serviceKey string) context.Context {
	h, _ := ctx.Value(identityKey{}).(headers)
	if h.Token == "" {
		h.Token = Incoming(ctx, "authorization")
	}
	if h.TOTP == "" {
		h.TOTP = Incoming(ctx, "x-totp-token")
	}
	if h.IP == "" {
		h.IP = Incoming(ctx, "x-client-ip")
	}
	if h.Agent == "" {
		h.Agent = Incoming(ctx, "x-client-agent")
	}
	return metadata.NewOutgoingContext(ctx, metadata.Pairs("x-service-key", serviceKey, "authorization", h.Token, "x-totp-token", h.TOTP, "x-client-ip", h.IP, "x-client-agent", h.Agent))
}

func ForwardIdentity(ctx context.Context, header http.Header) {
	identity, _ := ctx.Value(identityKey{}).(headers)
	if identity.Token != "" {
		header.Set("Authorization", "Bearer "+identity.Token)
	}
	if identity.IP != "" {
		header.Set("X-Real-IP", identity.IP)
	}
	if identity.Agent != "" {
		header.Set("User-Agent", identity.Agent)
	}
}

func RequireService(ctx context.Context, key string) error {
	got := Incoming(ctx, "x-service-key")
	if len(key) < 32 || subtle.ConstantTimeCompare([]byte(got), []byte(key)) != 1 {
		return status.Error(codes.Unauthenticated, "invalid service credentials")
	}
	return nil
}

func UnarySecurity(key string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, next grpc.UnaryHandler) (any, error) {
		if !strings.HasSuffix(info.FullMethod, "/Health") {
			if err := RequireService(ctx, key); err != nil {
				return nil, err
			}
		}
		return next(ctx, req)
	}
}

func StreamSecurity(key string) grpc.StreamServerInterceptor {
	return func(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, next grpc.StreamHandler) error {
		if err := RequireService(stream.Context(), key); err != nil {
			return err
		}
		return next(srv, stream)
	}
}
