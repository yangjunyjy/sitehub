package logic

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	"sitehub/service/blog/api/internal/svc"
	"sitehub/service/blog/api/internal/types"
)

type DeleteCategoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteCategoryLogic(ctx context.Context, s *svc.ServiceContext) *DeleteCategoryLogic {
	return &DeleteCategoryLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *DeleteCategoryLogic) DeleteCategory(req *types.DeleteCategoryRequest) (*types.Response, error) {

	_, err := l.svcCtx.Require(l.ctx, "blog:category:delete", "admin", true)
	if err != nil {
		return nil, err
	}

	err = l.svcCtx.Core.Categories.DeleteCategory(uint(req.ID))
	return done(err)
}
