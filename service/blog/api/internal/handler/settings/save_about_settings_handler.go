// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package settings

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"sitehub/service/blog/api/internal/logic/settings"
	"sitehub/service/blog/api/internal/svc"
	"sitehub/service/blog/api/internal/types"
)

func SaveAboutSettingsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AboutSettingsRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := settings.NewSaveAboutSettingsLogic(r.Context(), svcCtx)
		resp, err := l.SaveAboutSettings(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
