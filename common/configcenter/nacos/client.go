package nacos

import (
	"fmt"
	"sitehub/common/nacosconn"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

type Config struct {
	Enabled bool   `json:",default=false"`
	DataID  string `json:",optional"`
}

type Client struct {
	sdk           config_client.IConfigClient
	group, dataID string
}

func New(name string, connection nacosconn.Config, config Config) (*Client, error) {
	if config.DataID == "" {
		return nil, fmt.Errorf("ConfigCenter.DataID is required")
	}
	params, err := connection.Params(name, "config")
	if err != nil {
		return nil, err
	}
	sdk, err := clients.NewConfigClient(params)
	if err != nil {
		return nil, fmt.Errorf("connect to configuration center failed")
	}
	return &Client{sdk: sdk, group: connection.Group, dataID: config.DataID}, nil
}

func (c *Client) Get() ([]byte, error) {
	value, err := c.sdk.GetConfig(vo.ConfigParam{DataId: c.dataID, Group: c.group})
	if err != nil || value == "" {
		return nil, fmt.Errorf("required Nacos configuration is unavailable or empty")
	}
	return []byte(value), nil
}

func (c *Client) Watch(apply func([]byte)) error {
	return c.sdk.ListenConfig(vo.ConfigParam{DataId: c.dataID, Group: c.group,
		OnChange: func(_, _, _, data string) { apply([]byte(data)) }})
}

func (c *Client) Close() {
	_ = c.sdk.CancelListenConfig(vo.ConfigParam{DataId: c.dataID, Group: c.group})
	c.sdk.CloseClient()
}
