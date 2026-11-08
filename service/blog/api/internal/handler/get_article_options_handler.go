// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package handler

import (
	"net/http"

	"sitehub/service/blog/api/internal/logic"
	"sitehub/service/blog/api/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetArticleOptionsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewGetArticleOptionsLogic(r.Context(), svcCtx)
		resp, err := l.GetArticleOptions()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
