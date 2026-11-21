// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package statistics

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"sitehub/service/blog/api/internal/logic/statistics"
	"sitehub/service/blog/api/internal/svc"
)

func GetArticleStatsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := statistics.NewGetArticleStatsLogic(r.Context(), svcCtx)
		resp, err := l.GetArticleStats()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
