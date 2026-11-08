package logic

import (
	"context"
	"sitehub/service/blog/rpc/internal/svc"
	pb "sitehub/service/blog/rpc/pb/blog"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListCategoriesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListCategoriesLogic(ctx context.Context, s *svc.ServiceContext) *ListCategoriesLogic {
	return &ListCategoriesLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *ListCategoriesLogic) ListCategories(in *pb.CategoryListQueryRequest) (*pb.CategoryPage, error) {
	items, total, err := l.svcCtx.Core.Categories.ListCategories(int(in.Page), int(in.PageSize), in.Name)
	return reply[pb.CategoryPage](map[string]any{"items": items, "total": total}, err)
}
