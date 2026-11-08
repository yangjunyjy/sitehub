// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package handler

import (
	"net/http"

	"sitehub/service/product/api/internal/logic"
	"sitehub/service/product/api/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func DocsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewDocsLogic(r.Context(), svcCtx)
		resp, err := l.Docs()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
