package logic

import (
	"context"
	"sitehub/service/blog/rpc/internal/svc"
	pb "sitehub/service/blog/rpc/pb/blog"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteCategoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteCategoryLogic(ctx context.Context, s *svc.ServiceContext) *DeleteCategoryLogic {
	return &DeleteCategoryLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *DeleteCategoryLogic) DeleteCategory(in *pb.IDRequest) (*pb.Empty, error) {
	actor, err := l.svcCtx.Require(l.ctx, "blog:category:delete", "admin", true)
	if err != nil {
		return nil, err
	}
	_ = actor

	err = l.svcCtx.Core.Categories.DeleteCategory(uint(in.Id))
	return done(err)
}
