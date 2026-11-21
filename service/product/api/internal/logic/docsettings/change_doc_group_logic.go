package docsettings

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"sitehub/service/product/api/internal/logic"
	"sitehub/service/product/api/internal/svc"
	"sitehub/service/product/api/internal/types"
	"sitehub/service/product/domain"
)

type ChangeDocGroupLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewChangeDocGroupLogic(ctx context.Context, s *svc.ServiceContext) *ChangeDocGroupLogic {
	return &ChangeDocGroupLogic{ctx, s, logx.WithContext(ctx)}
}

func (l *ChangeDocGroupLogic) ChangeDocGroup(req *types.ChangeDocGroupRequest) (*types.Response, error) {
	_, err := l.svcCtx.Require(l.ctx, "website:docs:edit", "admin", true)
	if err != nil {
		return nil, err
	}
	input := domain.DocGroupInput{
		Action:    req.Action,
		Version:   req.Version,
		Title:     req.Title,
		ID:        req.ID,
		Revision:  req.Revision,
		Direction: int(req.Direction),
	}

	err = l.svcCtx.Core.Content.ChangeDocGroup(input)
	return logic.Response(struct{}{}, err)
}
