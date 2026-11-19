package etcd

import (
	"bytes"
	"context"
	"fmt"
	"strings"
)

type ConfigCenter struct{ *base }

func NewConfigCenter(options Options) (*ConfigCenter, error) {
	b, err := connect(options)
	if err != nil {
		return nil, err
	}
	return &ConfigCenter{base: b}, nil
}

func validKey(key string) bool {
	return key != "" && key != "." && key != ".." && !strings.ContainsAny(key, "/\\\x00")
}

func (c *ConfigCenter) Get(ctx context.Context, key string) ([]byte, error) {
	if !validKey(key) {
		return nil, fmt.Errorf("configuration key must be a non-empty path segment")
	}
	if err := c.begin(); err != nil {
		return nil, err
	}
	defer c.wg.Done()
	response, err := c.get(ctx, c.options.Prefix+key, false)
	if err != nil {
		return nil, err
	}
	if len(response.Kvs) == 0 || len(response.Kvs[0].Value) == 0 {
		return nil, fmt.Errorf("required configuration is unavailable or empty")
	}
	return bytes.Clone(response.Kvs[0].Value), nil
}

func (c *ConfigCenter) Watch(ctx context.Context, key string, update func([]byte, error)) (func(), error) {
	if !validKey(key) || update == nil {
		return nil, fmt.Errorf("configuration key and watch callback are required")
	}
	fullKey := c.options.Prefix + key
	return c.watch(ctx, fullKey, false, func(values map[string][]byte, err error) { update(bytes.Clone(values[fullKey]), err) })
}

func (c *ConfigCenter) Close() error { return c.close() }
