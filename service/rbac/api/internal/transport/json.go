package transport

import (
	"context"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"sitehub/service/rbac/api/internal/svc"
	"sitehub/service/rbac/api/internal/types"
)

func WithRequest[T any](services *svc.ServiceContext, action func(context.Context, *svc.ServiceContext, *T) (*types.Response, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req T
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, status.Error(codes.InvalidArgument, "invalid request"))
			return
		}
		resp, err := action(r.Context(), services, &req)
		write(w, r, resp, err)
	}
}

func WithoutRequest(services *svc.ServiceContext, action func(context.Context, *svc.ServiceContext) (*types.Response, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := action(r.Context(), services)
		write(w, r, resp, err)
	}
}

func write(w http.ResponseWriter, r *http.Request, resp *types.Response, err error) {
	if err != nil {
		httpx.ErrorCtx(r.Context(), w, err)
		return
	}
	httpx.OkJsonCtx(r.Context(), w, resp)
}
