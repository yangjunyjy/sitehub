package storage

import (
	"net/http"
	"sitehub/service/blog/api/internal/svc"
)

func EnsureStorageDefaultsHandler(services *svc.ServiceContext) http.HandlerFunc {
	return internalRequest(services, func(*http.Request) (any, error) { return nil, services.Core.StorageSettings.EnsureDefaults() })
}
