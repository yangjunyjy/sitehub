package categories

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	logicshared "sitehub/service/blog/api/internal/logic"
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
	return logicshared.ReplyAs[payload.CategoryPage](map[string]any{"items": items}, err)
}
