package review

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	logicshared "sitehub/service/blog/api/internal/logic"
	"sitehub/service/blog/api/internal/svc"
	"sitehub/service/blog/api/internal/types"
)

type CommunityReviewQueueLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCommunityReviewQueueLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CommunityReviewQueueLogic {
	return &CommunityReviewQueueLogic{ctx, svcCtx}
}

func validQueue(kind string) bool {
	return kind == "friends" || kind == "messages" || kind == "comments"
}

func (l *CommunityReviewQueueLogic) CommunityReviewQueue(req *types.ReviewQueue) (*types.Response, error) {
	if _, err := l.svcCtx.Require(l.ctx, "blog:community:list", "admin", false); err != nil {
		return nil, err
	}
	if req.Page == 0 {
		req.Page = 1
	}
	if !validQueue(req.Kind) || req.Page < 1 || req.Page > 1000000 || (req.Status != "approved" && req.Status != "pending" && req.Status != "rejected") {
		return nil, status.Error(codes.InvalidArgument, "审核列表参数无效")
	}
	items, total, err := l.svcCtx.Core.Community.ReviewQueue(l.ctx, req.Kind, req.Status, req.Page)
	return logicshared.Reply(map[string]any{"items": items, "total": total}, err)
}
