package handler

import (
	"context"
	"github.com/zeromicro/go-zero/rest/httpx"
	"net/http"
	"sitehub/service/blog/api/internal/logic"
	"sitehub/service/blog/api/internal/svc"
	"sitehub/service/blog/api/internal/types"
)

func reviewHandler[Request any](services *svc.ServiceContext, action func(context.Context, *svc.ServiceContext, *Request) (*types.Response, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req Request
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		resp, err := action(r.Context(), services, &req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

func CommunityReviewQueueHandler(services *svc.ServiceContext) http.HandlerFunc {
	return reviewHandler(services, logic.CommunityReviewQueue)
}
func ReviewCommunityHandler(services *svc.ServiceContext) http.HandlerFunc {
	return reviewHandler(services, logic.ReviewCommunity)
}
