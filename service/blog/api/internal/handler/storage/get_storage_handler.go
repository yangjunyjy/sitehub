// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package storage

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"sitehub/service/blog/api/internal/logic/storage"
	"sitehub/service/blog/api/internal/svc"
)

func GetStorageHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := storage.NewGetStorageLogic(r.Context(), svcCtx)
		resp, err := l.GetStorage()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
