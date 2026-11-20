package logic

import (
	"context"
	"errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sitehub/service/blog/api/internal/svc"
	"sitehub/service/blog/api/internal/types"
	"sitehub/service/blog/domain"
)

func validQueue(kind string) bool {
	return kind == "friends" || kind == "messages" || kind == "comments"
}

func CommunityReviewQueue(ctx context.Context, services *svc.ServiceContext, req *types.ReviewQueue) (*types.Response, error) {
	if _, err := services.Require(ctx, "blog:community:list", "admin", false); err != nil {
		return nil, err
	}
	if req.Page == 0 {
		req.Page = 1
	}
	if !validQueue(req.Kind) || req.Page < 1 || req.Page > 1000000 || (req.Status != "approved" && req.Status != "pending" && req.Status != "rejected") {
		return nil, status.Error(codes.InvalidArgument, "审核列表参数无效")
	}
	items, total, err := services.Core.Community.ReviewQueue(ctx, req.Kind, req.Status, req.Page)
	return reply(map[string]any{"items": items, "total": total}, err)
}

func ReviewCommunity(ctx context.Context, services *svc.ServiceContext, req *types.ReviewDecision) (*types.Response, error) {
	if _, err := services.Require(ctx, "blog:community:review", "admin", true); err != nil {
		return nil, err
	}
	if !validQueue(req.Kind) || req.ID == 0 || (req.Status != "approved" && req.Status != "rejected") {
		return nil, status.Error(codes.InvalidArgument, "审核操作无效")
	}
	err := services.Core.Community.Review(ctx, req.Kind, req.ID, req.Status)
	if errors.Is(err, domain.ErrCommunityItemNotFound) {
		return nil, status.Error(codes.NotFound, "记录不存在")
	}
	return done(err)
}
