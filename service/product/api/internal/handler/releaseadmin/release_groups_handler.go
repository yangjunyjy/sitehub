// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package releaseadmin

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"sitehub/service/product/api/internal/logic/releaseadmin"
	"sitehub/service/product/api/internal/svc"
)

func ReleaseGroupsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := releaseadmin.NewReleaseGroupsLogic(r.Context(), svcCtx)
		resp, err := l.ReleaseGroups()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
