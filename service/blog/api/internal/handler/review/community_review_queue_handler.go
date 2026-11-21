package review

import (
	"github.com/zeromicro/go-zero/rest/httpx"
	"net/http"
	"sitehub/service/blog/api/internal/logic/review"
	"sitehub/service/blog/api/internal/svc"
	"sitehub/service/blog/api/internal/types"
)

func CommunityReviewQueueHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ReviewQueue
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		resp, err := review.NewCommunityReviewQueueLogic(r.Context(), svcCtx).CommunityReviewQueue(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}
