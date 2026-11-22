package manifest

import (
	"fmt"
	"net"
	"net/netip"
	"regexp"
	"strconv"
	"strings"
)

// HostMatcher represents the optional host and port of an HTTP route.
// An empty Name matches every host. An empty Port matches every port.
type HostMatcher struct {
	Name string
	Port string
}

var dnsLabel = regexp.MustCompile(`^[a-z0-9_]([a-z0-9_-]{0,61}[a-z0-9_])?$`)

func ParseHost(value string) (HostMatcher, error) {
	if value == "" || value == "*" {
		return HostMatcher{}, nil
	}
	invalid := func() error {
		return fmt.Errorf("expected DNS name, IP address or * with optional port, got %q", value)
	}
	if addr, err := netip.ParseAddr(value); err == nil && addr.Zone() == "" {
		return HostMatcher{Name: addr.String()}, nil
	}
	host, port := value, ""
	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		host = host[1 : len(host)-1]
	} else if strings.Contains(host, ":") {
		var err error
		host, port, err = net.SplitHostPort(value)
		if err != nil {
			return HostMatcher{}, invalid()
		}
		number, err := strconv.ParseUint(port, 10, 16)
		if err != nil || number == 0 {
			return HostMatcher{}, invalid()
		}
		port = strconv.FormatUint(number, 10)
	}
	if addr, err := netip.ParseAddr(host); err == nil && addr.Zone() == "" {
		return HostMatcher{Name: addr.String(), Port: port}, nil
	}
	if strings.HasPrefix(value, "[") {
		return HostMatcher{}, invalid()
	}
	if host == "*" {
		return HostMatcher{Port: port}, nil
	}
	host = strings.ToLower(strings.TrimSuffix(host, "."))
	if len(host) > 253 {
		return HostMatcher{}, invalid()
	}
	for _, label := range strings.Split(host, ".") {
		if !dnsLabel.MatchString(label) {
			return HostMatcher{}, invalid()
		}
	}
	return HostMatcher{Name: host, Port: port}, nil
}

func (m HostMatcher) Matches(request HostMatcher) bool {
	return (m.Name == "" || m.Name == request.Name) && (m.Port == "" || m.Port == request.Port)
}

func (m HostMatcher) Specificity() int {
	score := 0
	if m.Name != "" {
		score += 2
	}
	if m.Port != "" {
		score++
	}
	return score
}
