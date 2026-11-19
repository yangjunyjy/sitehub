package logic

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	"sitehub/service/blog/api/internal/payload"
	"sitehub/service/blog/api/internal/svc"
	"sitehub/service/blog/api/internal/types"
)

type ListCategoriesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListCategoriesLogic(ctx context.Context, s *svc.ServiceContext) *ListCategoriesLogic {
	return &ListCategoriesLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *ListCategoriesLogic) ListCategories(req *types.ListCategoriesRequest) (*types.Response, error) {

	items, total, err := l.svcCtx.Core.Categories.ListCategories(int(req.Page), int(req.PageSize), req.Name)
	return ReplyAs[payload.CategoryPage](map[string]any{"items": items, "total": total}, err)
}
