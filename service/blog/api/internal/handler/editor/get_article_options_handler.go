// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package editor

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"sitehub/service/blog/api/internal/logic/editor"
	"sitehub/service/blog/api/internal/svc"
)

func GetArticleOptionsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := editor.NewGetArticleOptionsLogic(r.Context(), svcCtx)
		resp, err := l.GetArticleOptions()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
