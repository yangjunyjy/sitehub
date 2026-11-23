// Package clientaddr resolves client identity at the public HTTP boundary.
package clientaddr

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"strings"
)

type Trusted []netip.Prefix

func Parse(cidrs []string) (Trusted, error) {
	var nets Trusted
	for _, cidr := range cidrs {
		prefix, err := netip.ParsePrefix(cidr)
		if err != nil {
			return nil, fmt.Errorf("invalid HTTP.TrustedProxies CIDR %q: %w", cidr, err)
		}
		nets = append(nets, prefix)
	}
	return nets, nil
}

func (nets Trusted) contains(addr netip.Addr) bool {
	for _, prefix := range nets {
		if prefix.Contains(addr.Unmap()) {
			return true
		}
	}
	return false
}

type Info struct {
	Client string
	Scheme string
	Host   string
}

type infoKey struct{}

func (info Info) Context(ctx context.Context) context.Context {
	return context.WithValue(ctx, infoKey{}, info)
}

func From(r *http.Request) Info {
	if info, ok := r.Context().Value(infoKey{}).(Info); ok {
		return info
	}
	return Trusted(nil).Resolve(r)
}

func (nets Trusted) Resolve(r *http.Request) Info {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	info := Info{Client: host, Scheme: "http", Host: r.Host}
	if r.TLS != nil {
		info.Scheme = "https"
	}
	peer, err := netip.ParseAddr(host)
	if err != nil || !nets.contains(peer) {
		return info
	}
	if proto := r.Header.Values("X-Forwarded-Proto"); len(proto) == 1 && (proto[0] == "http" || proto[0] == "https") {
		info.Scheme = proto[0]
	}
	chain := r.Header.Values("X-Forwarded-For")
	if len(chain) == 0 {
		real := r.Header.Values("X-Real-IP")
		if len(real) == 1 {
			if ip, err := netip.ParseAddr(strings.TrimSpace(real[0])); err == nil && ip.Zone() == "" {
				info.Client = ip.Unmap().String()
			}
		}
		return info
	}
	// Follow the chain from the trusted socket. A malformed entry terminates the chain.
	for i := len(chain) - 1; i >= 0; i-- {
		hops := strings.Split(chain[i], ",")
		for j := len(hops) - 1; j >= 0; j-- {
			ip, err := netip.ParseAddr(strings.TrimSpace(hops[j]))
			if err != nil || ip.Zone() != "" {
				return info
			}
			info.Client = ip.Unmap().String()
			if !nets.contains(ip) {
				return info
			}
		}
	}
	return info
}
