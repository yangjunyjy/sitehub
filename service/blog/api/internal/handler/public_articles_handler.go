// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package handler

import (
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"sitehub/service/blog/api/internal/logic"
	"sitehub/service/blog/api/internal/svc"
	"sitehub/service/blog/api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func PublicArticlesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PublicArticlesRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, status.Error(codes.InvalidArgument, "invalid request"))
			return
		}

		l := logic.NewPublicArticlesLogic(r.Context(), svcCtx)
		resp, err := l.PublicArticles(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
