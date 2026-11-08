package logic

import (
	"context"
	"sitehub/service/blog/rpc/internal/svc"
	pb "sitehub/service/blog/rpc/pb/blog"

	"github.com/zeromicro/go-zero/core/logx"
)

type CategoryTreeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCategoryTreeLogic(ctx context.Context, s *svc.ServiceContext) *CategoryTreeLogic {
	return &CategoryTreeLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *CategoryTreeLogic) CategoryTree(in *pb.Empty) (*pb.CategoryPage, error) {
	items, err := l.svcCtx.Core.Categories.GetCategoryTree()
	return reply[pb.CategoryPage](map[string]any{"items": items}, err)
}
