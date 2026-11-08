// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package handler

import (
	"net/http"

	"sitehub/service/rbac/api/internal/logic"
	"sitehub/service/rbac/api/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func CurrentMemberHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewCurrentMemberLogic(r.Context(), svcCtx)
		resp, err := l.CurrentMember()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
