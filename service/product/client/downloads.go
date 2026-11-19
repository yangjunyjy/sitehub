package client

import (
	"context"
	"net/http"
	"net/url"
	"sitehub/common/servicehttp"
	"sitehub/service/product/contract"
)

type Downloads struct{ baseURL, key string }

func New(baseURL, key string) *Downloads { return &Downloads{baseURL: baseURL, key: key} }
func (d *Downloads) ResolveDownload(ctx context.Context, id string) (contract.ProductDownload, error) {
	var result contract.ProductDownload
	err := servicehttp.Do(ctx, d.baseURL, d.key, http.MethodGet, "/internal/product/download?id="+url.QueryEscape(id), nil, &result)
	return result, err
}
