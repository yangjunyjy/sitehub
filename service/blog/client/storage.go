package client

import (
	"context"
	"net/http"
	"sitehub/common/servicehttp"
	"sitehub/service/blog/contract"
)

type Storage struct{ baseURL, key string }

func New(baseURL, key string) *Storage { return &Storage{baseURL: baseURL, key: key} }
func (s *Storage) EnsureDefaults() error {
	return servicehttp.Do(context.Background(), s.baseURL, s.key, http.MethodPost, "/internal/blog/storage/defaults", nil, nil)
}
func (s *Storage) GetStorage() (contract.WebsiteStorage, error) {
	var result contract.WebsiteStorage
	err := servicehttp.Do(context.Background(), s.baseURL, s.key, http.MethodGet, "/internal/blog/storage", nil, &result)
	return result, err
}
func (s *Storage) SaveStorage(value contract.WebsiteStorage) error {
	return servicehttp.Do(context.Background(), s.baseURL, s.key, http.MethodPut, "/internal/blog/storage", value, nil)
}
func (s *Storage) ValidateRelease(location string) error {
	return servicehttp.Do(context.Background(), s.baseURL, s.key, http.MethodPost, "/internal/blog/storage/validate", map[string]string{"location": location}, nil)
}
