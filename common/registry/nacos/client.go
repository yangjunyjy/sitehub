package nacos

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"sitehub/common/nacosconn"
)

type Config struct {
	Enabled        bool          `json:",default=false"`
	AdvertiseIP    string        `json:",optional"`
	AllowLoopback  bool          `json:",default=false"`
	MaxStale       time.Duration `json:",default=30s"`
	StartupTimeout time.Duration `json:",default=30s"`
}
type Client struct {
	naming_client.INamingClient
	group string
}

func New(name string, connection nacosconn.Config) (*Client, error) {
	params, err := connection.Params(name, "naming")
	if err != nil {
		return nil, err
	}
	sdk, err := clients.NewNamingClient(params)
	if err != nil {
		return nil, fmt.Errorf("connect to registry failed")
	}
	return &Client{INamingClient: sdk, group: connection.Group}, nil
}

// ResolveAdvertise selects a routable address, never an unspecified listener.
// Container hostnames resolve to each container's own IP, permitting replicas.
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

func (c *Client) Register(name, ip string, port uint64) (func() error, error) {
	protocol := "http"
	if strings.HasSuffix(name, "-rpc") {
		protocol = "grpc"
	}
	ok, err := c.RegisterInstance(vo.RegisterInstanceParam{Ip: ip, Port: port, Weight: 100, Enable: true, Healthy: true,
		ServiceName: name, GroupName: c.group, ClusterName: "DEFAULT", Ephemeral: true, Metadata: map[string]string{"protocol": protocol}})
	if err != nil || !ok {
		return nil, fmt.Errorf("Nacos registration failed for %s", name)
	}
	return func() error {
		ok, err := c.DeregisterInstance(vo.DeregisterInstanceParam{Ip: ip, Port: port, ServiceName: name, GroupName: c.group, Cluster: "DEFAULT", Ephemeral: true})
		if err != nil || !ok {
			return fmt.Errorf("Nacos deregistration failed for %s", name)
		}
		return nil
	}, nil
}
