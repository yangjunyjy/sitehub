// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package community

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"sitehub/service/blog/api/internal/logic/community"
	"sitehub/service/blog/api/internal/svc"
)

func ListFriendsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := community.NewListFriendsLogic(r.Context(), svcCtx)
		resp, err := l.ListFriends()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
