// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package documents

import (
	"net/http"

	"sitehub/service/product/api/internal/logic/documents"
	"sitehub/service/product/api/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func DocModesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := documents.NewDocModesLogic(r.Context(), svcCtx)
		resp, err := l.DocModes()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
