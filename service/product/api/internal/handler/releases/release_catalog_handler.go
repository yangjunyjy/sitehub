// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package releases

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"sitehub/service/product/api/internal/logic/releases"
	"sitehub/service/product/api/internal/svc"
)

func ReleaseCatalogHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := releases.NewReleaseCatalogLogic(r.Context(), svcCtx)
		resp, err := l.ReleaseCatalog()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
