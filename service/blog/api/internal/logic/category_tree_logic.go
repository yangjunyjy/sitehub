package logic

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	"sitehub/service/blog/api/internal/payload"
	"sitehub/service/blog/api/internal/svc"
	"sitehub/service/blog/api/internal/types"
)

type CategoryTreeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCategoryTreeLogic(ctx context.Context, s *svc.ServiceContext) *CategoryTreeLogic {
	return &CategoryTreeLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *CategoryTreeLogic) CategoryTree() (*types.Response, error) {

	items, err := l.svcCtx.Core.Categories.GetCategoryTree()
	return ReplyAs[payload.CategoryPage](map[string]any{"items": items}, err)
}
