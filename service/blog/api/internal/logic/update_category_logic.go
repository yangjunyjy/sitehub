package logic

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	"sitehub/service/blog/api/internal/svc"
	"sitehub/service/blog/api/internal/types"
	entity "sitehub/service/blog/model/entity"
)

type UpdateCategoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateCategoryLogic(ctx context.Context, s *svc.ServiceContext) *UpdateCategoryLogic {
	return &UpdateCategoryLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *UpdateCategoryLogic) UpdateCategory(req *types.UpdateCategoryRequest) (*types.Response, error) {

	_, err := l.svcCtx.Require(l.ctx, "blog:category:update", "admin", true)
	if err != nil {
		return nil, err
	}
	input := entity.Category{ID: uint(req.ID), ParentID: req.ParentID, SortOrder: int(req.SortOrder), Name: req.Name, Slug: req.Slug, Description: textOrEmpty(req.Description), Icon: textOrEmpty(req.Icon)}

	err = l.svcCtx.Core.Categories.UpdateCategory(uint(req.ID), &input)
	return done(err)
}
