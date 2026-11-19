package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"sitehub/service/product/api/internal/svc"
	"sitehub/service/product/api/internal/types"
	"sitehub/service/product/domain"
)

type ChangeDocVersionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewChangeDocVersionLogic(ctx context.Context, s *svc.ServiceContext) *ChangeDocVersionLogic {
	return &ChangeDocVersionLogic{ctx, s, logx.WithContext(ctx)}
}

func (l *ChangeDocVersionLogic) ChangeDocVersion(req *types.ChangeDocVersionRequest) (*types.Response, error) {
	actor, err := l.svcCtx.Require(l.ctx, "website:docs:edit", "admin", true)
	if err != nil {
		return nil, err
	}
	input := domain.DocVersionInput{
		Action:   req.Action,
		ID:       req.ID,
		Name:     req.Name,
		Status:   req.Status,
		Source:   req.Source,
		ModeID:   req.ModeID,
		Revision: req.Revision,
	}

	err = l.svcCtx.Core.Content.ChangeDocVersion(input, actor.Id)
	return response(struct{}{}, err)
}
