package storage

import (
	"encoding/json"
	"net/http"
	"sitehub/common/errorx"
	"sitehub/service/blog/api/internal/svc"
	"sitehub/service/blog/api/internal/types"
)

func ValidateReleaseHandler(services *svc.ServiceContext) http.HandlerFunc {
	return internalRequest(services, func(r *http.Request) (any, error) {
		var input types.ValidateReleaseRequest
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			return nil, errorx.Invalid("invalid release location")
		}
		return nil, services.Core.StorageSettings.ValidateRelease(input.Location)
	})
}
