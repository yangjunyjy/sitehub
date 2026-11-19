package handler

import (
	"encoding/json"
	"github.com/zeromicro/go-zero/rest"
	"net/http"
	"sitehub/common/errorx"
	"sitehub/common/servicehttp"
	"sitehub/service/blog/api/internal/svc"
	"sitehub/service/blog/contract"
)

func RegisterStorageHTTP(server *rest.Server, services *svc.ServiceContext) {
	register := func(method, path string, handle func(*http.Request) (any, error)) {
		server.AddRoute(rest.Route{Method: method, Path: path, Handler: func(w http.ResponseWriter, r *http.Request) {
			if !servicehttp.Authorized(r, services.Config.ServiceKey) {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
			value, err := handle(r)
			servicehttp.Write(w, value, err)
		}})
	}
	register(http.MethodGet, "/internal/blog/storage", func(*http.Request) (any, error) { return services.Core.StorageSettings.GetStorage() })
	register(http.MethodPost, "/internal/blog/storage/defaults", func(*http.Request) (any, error) { return nil, services.Core.StorageSettings.EnsureDefaults() })
	register(http.MethodPut, "/internal/blog/storage", func(r *http.Request) (any, error) {
		var value contract.WebsiteStorage
		if err := json.NewDecoder(r.Body).Decode(&value); err != nil {
			return nil, errorx.Invalid("invalid storage settings")
		}
		return nil, services.Core.StorageSettings.SaveStorage(value)
	})
	register(http.MethodPost, "/internal/blog/storage/validate", func(r *http.Request) (any, error) {
		var input struct {
			Location string `json:"location"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			return nil, errorx.Invalid("invalid release location")
		}
		return nil, services.Core.StorageSettings.ValidateRelease(input.Location)
	})
}
