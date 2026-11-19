package logic

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	"sitehub/service/blog/api/internal/payload"
	"sitehub/service/blog/api/internal/svc"
	"sitehub/service/blog/api/internal/types"
)

type GetCategoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCategoryLogic(ctx context.Context, s *svc.ServiceContext) *GetCategoryLogic {
	return &GetCategoryLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *GetCategoryLogic) GetCategory(req *types.GetCategoryRequest) (*types.Response, error) {

	value, err := l.svcCtx.Core.Categories.GetCategory(uint(req.ID))
	return ReplyAs[payload.Category](value, err)
}
