// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package handler

import (
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"sitehub/service/file/api/internal/logic"
	"sitehub/service/file/api/internal/svc"
	"sitehub/service/file/api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func SaveStorageHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SaveStorageRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, status.Error(codes.InvalidArgument, "invalid request"))
			return
		}

		l := logic.NewSaveStorageLogic(r.Context(), svcCtx)
		resp, err := l.SaveStorage(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
