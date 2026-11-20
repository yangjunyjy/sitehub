package handler

import (
	"github.com/zeromicro/go-zero/rest/httpx"
	"net/http"
	"sitehub/service/blog/api/internal/logic"
	"sitehub/service/blog/api/internal/svc"
	"sitehub/service/blog/api/internal/types"
)

func SaveHomeSettingsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.HomeSettingsRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		resp, err := logic.NewSaveHomeSettingsLogic(r.Context(), svcCtx).SaveHomeSettings(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}
