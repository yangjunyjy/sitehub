package storage

import (
	"net/http"
	"sitehub/service/blog/api/internal/svc"
)

func ReadInternalStorageHandler(services *svc.ServiceContext) http.HandlerFunc {
	return internalRequest(services, func(*http.Request) (any, error) { return services.Core.StorageSettings.GetStorage() })
}
