package downloads

import (
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"sitehub/common/servicehttp"
	"sitehub/service/product/api/internal/svc"
)

func InternalDownloadHandler(service *svc.ServiceContext) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if !servicehttp.Authorized(request, service.Config.ServiceKey) {
			servicehttp.Write(writer, nil, status.Error(codes.Unauthenticated, "invalid service credentials"))
			return
		}
		if err := service.CheckClient(request.Context()); err != nil {
			servicehttp.Write(writer, nil, err)
			return
		}
		if _, err := service.Require(request.Context(), "", "member", false); err != nil {
			servicehttp.Write(writer, nil, err)
			return
		}
		value, err := service.Core.Releases.Resolve(request.Context(), request.URL.Query().Get("id"))
		servicehttp.Write(writer, value, err)
	}
}
