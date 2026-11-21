// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package downloads

import (
	"net/http"

	"sitehub/service/product/api/internal/logic/downloads"
	"sitehub/service/product/api/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func DownloadsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := downloads.NewDownloadsLogic(r.Context(), svcCtx)
		resp, err := l.Downloads()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
