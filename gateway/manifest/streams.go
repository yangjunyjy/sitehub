package manifest

import (
	"fmt"
	"net"
	"net/netip"
	"regexp"
	"strconv"
	"strings"
)

// Stream forwards one TCP listener or UDP packet socket to a network endpoint.
type Stream struct {
	ID       string `json:"id"`
	Protocol string `json:"protocol"`
	Listen   string `json:"listen"`
	Target   string `json:"target"`
}

var streamDNSLabel = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?$`)

func validStreamAddress(address string, listen bool) bool {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return false
	}
	p, err := strconv.ParseUint(port, 10, 16)
	if err != nil || p == 0 || strconv.FormatUint(p, 10) != port {
		return false
	}
	if host == "" {
		return listen
	}
	if _, err := netip.ParseAddr(host); err == nil {
		return true
	}
	if len(host) > 253 {
		return false
	}
	for _, label := range strings.Split(strings.TrimSuffix(host, "."), ".") {
		if !streamDNSLabel.MatchString(label) {
			return false
		}
	}
	return true
}

func validateStreams(streams []Stream) error {
	ids := map[string]bool{}
	listeners := map[string]bool{}
	for _, stream := range streams {
		if !identifier.MatchString(stream.ID) || ids[stream.ID] {
			return fmt.Errorf("invalid or duplicate stream id %q", stream.ID)
		}
		ids[stream.ID] = true
		if stream.Protocol != "tcp" && stream.Protocol != "udp" {
			return fmt.Errorf("stream %s: protocol must be tcp or udp", stream.ID)
		}
		if !validStreamAddress(stream.Listen, true) || !validStreamAddress(stream.Target, false) {
			return fmt.Errorf("stream %s: use host:port for target and [host]:port for listen (port 1..65535)", stream.ID)
		}
		key := stream.Protocol + "/" + strings.ToLower(stream.Listen)
		if listeners[key] {
			return fmt.Errorf("stream %s: duplicate listener %s", stream.ID, key)
		}
		listeners[key] = true
	}
	return nil
}
