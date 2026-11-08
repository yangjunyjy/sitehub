package logic

import (
	"context"
	"sitehub/service/blog/rpc/internal/svc"
	pb "sitehub/service/blog/rpc/pb/blog"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCategoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCategoryLogic(ctx context.Context, s *svc.ServiceContext) *GetCategoryLogic {
	return &GetCategoryLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *GetCategoryLogic) GetCategory(in *pb.IDRequest) (*pb.EntityCategory, error) {
	value, err := l.svcCtx.Core.Categories.GetCategory(uint(in.Id))
	return reply[pb.EntityCategory](value, err)
}
