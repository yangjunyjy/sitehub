package storage

import (
	"net/http"
	"sitehub/common/servicehttp"
	"sitehub/service/blog/api/internal/svc"
)

func internalRequest(services *svc.ServiceContext, handle func(*http.Request) (any, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !servicehttp.Authorized(r, services.Config.ServiceKey) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		value, err := handle(r)
		servicehttp.Write(w, value, err)
	}
}
