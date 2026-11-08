// Package nacosconn holds local bootstrap settings shared by the two Nacos clients.
package nacosconn

import (
	"fmt"
	"net"
	"path/filepath"
	"strconv"

	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

type Config struct {
	Servers     []string `json:",optional"`
	NamespaceID string   `json:",optional"`
	Group       string   `json:",default=SITEHUB"`
	Username    string   `json:",optional"`
	Password    string   `json:",optional"`
	TimeoutMs   uint64   `json:",default=5000"`
	StateDir    string   `json:",default=.tmp/nacos"`
}

func (c Config) Params(name, kind string) (vo.NacosClientParam, error) {
	if len(c.Servers) == 0 || c.NamespaceID == "" || c.Group == "" {
		return vo.NacosClientParam{}, fmt.Errorf("Nacos Servers, NamespaceID and Group are required")
	}
	servers := make([]constant.ServerConfig, 0, len(c.Servers))
	for _, address := range c.Servers {
		host, port, err := net.SplitHostPort(address)
		n, parseErr := strconv.ParseUint(port, 10, 16)
		if err != nil || parseErr != nil || host == "" || n == 0 {
			return vo.NacosClientParam{}, fmt.Errorf("Nacos Servers must contain host:port addresses")
		}
		servers = append(servers, constant.ServerConfig{IpAddr: host, Port: n, Scheme: "http", ContextPath: "/nacos"})
	}
	return vo.NacosClientParam{ServerConfigs: servers, ClientConfig: &constant.ClientConfig{
		NamespaceId: c.NamespaceID, AppName: name, Username: c.Username, Password: c.Password,
		TimeoutMs: c.TimeoutMs, NotLoadCacheAtStart: true, UpdateCacheWhenEmpty: true,
		DisableUseSnapShot: true, LogLevel: "warn", CacheDir: filepath.Join(c.StateDir, name, kind, "cache"),
		LogDir: filepath.Join(c.StateDir, name, kind, "log"),
	}}, nil
}
