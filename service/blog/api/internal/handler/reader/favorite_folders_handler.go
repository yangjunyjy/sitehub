// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package reader

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"sitehub/service/blog/api/internal/logic/reader"
	"sitehub/service/blog/api/internal/svc"
)

func FavoriteFoldersHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := reader.NewFavoriteFoldersLogic(r.Context(), svcCtx)
		resp, err := l.FavoriteFolders()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
