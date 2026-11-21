// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package manualadmin

import (
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/zeromicro/go-zero/rest/httpx"
	"sitehub/service/product/api/internal/logic/manualadmin"
	"sitehub/service/product/api/internal/svc"
	"sitehub/service/product/api/internal/types"
)

func ManualModeDefaultHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ManualNamedRevision
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, status.Error(codes.InvalidArgument, "请求参数无效"))
			return
		}

		l := manualadmin.NewManualModeDefaultLogic(r.Context(), svcCtx)
		resp, err := l.ManualModeDefault(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
