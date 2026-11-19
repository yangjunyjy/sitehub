package registry

import (
	"fmt"
	"net"
	"strconv"
)

// ResolveAdvertise resolves container hostnames to their routable IPv4 address.
func ResolveAdvertise(advertise, listen string, allowLoopback bool) (string, uint64, error) {
	host, port, err := net.SplitHostPort(listen)
	if err != nil {
		return "", 0, fmt.Errorf("invalid service listen address")
	}
	n, err := strconv.ParseUint(port, 10, 16)
	if err != nil || n == 0 {
		return "", 0, fmt.Errorf("invalid service port")
	}
	if advertise != "" {
		host = advertise
	}
	ips, err := net.LookupIP(host)
	if err != nil {
		return "", 0, fmt.Errorf("cannot resolve Registry.AdvertiseIP")
	}
	for _, ip := range ips {
		if ip.IsUnspecified() || ip.IsMulticast() || (!allowLoopback && ip.IsLoopback()) {
			continue
		}
		if ip.To4() != nil {
			return ip.String(), n, nil
		}
	}
	return "", 0, fmt.Errorf("Registry.AdvertiseIP must resolve to a reachable IPv4 address")
}
