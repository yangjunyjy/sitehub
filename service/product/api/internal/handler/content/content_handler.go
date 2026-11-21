// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package content

import (
	"net/http"

	"sitehub/service/product/api/internal/logic/content"
	"sitehub/service/product/api/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func ContentHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := content.NewContentLogic(r.Context(), svcCtx)
		resp, err := l.Content()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
