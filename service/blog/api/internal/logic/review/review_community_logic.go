package review

import (
	"context"
	"errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	logicshared "sitehub/service/blog/api/internal/logic"
	"sitehub/service/blog/api/internal/svc"
	"sitehub/service/blog/api/internal/types"
	"sitehub/service/blog/domain"
)

type ReviewCommunityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReviewCommunityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReviewCommunityLogic {
	return &ReviewCommunityLogic{ctx, svcCtx}
}

func (l *ReviewCommunityLogic) ReviewCommunity(req *types.ReviewDecision) (*types.Response, error) {
	if _, err := l.svcCtx.Require(l.ctx, "blog:community:review", "admin", true); err != nil {
		return nil, err
	}
	if !validQueue(req.Kind) || req.ID == 0 || (req.Status != "approved" && req.Status != "rejected") {
		return nil, status.Error(codes.InvalidArgument, "审核操作无效")
	}
	err := l.svcCtx.Core.Community.Review(l.ctx, req.Kind, req.ID, req.Status)
	if errors.Is(err, domain.ErrCommunityItemNotFound) {
		return nil, status.Error(codes.NotFound, "记录不存在")
	}
	return logicshared.Done(err)
}
