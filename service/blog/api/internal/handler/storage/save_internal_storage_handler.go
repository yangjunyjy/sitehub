package storage

import (
	"encoding/json"
	"net/http"
	"sitehub/common/errorx"
	"sitehub/service/blog/api/internal/svc"
	"sitehub/service/blog/contract"
)

func SaveInternalStorageHandler(services *svc.ServiceContext) http.HandlerFunc {
	return internalRequest(services, func(r *http.Request) (any, error) {
		var value contract.WebsiteStorage
		if err := json.NewDecoder(r.Body).Decode(&value); err != nil {
			return nil, errorx.Invalid("invalid storage settings")
		}
		return nil, services.Core.StorageSettings.SaveStorage(value)
	})
}
