package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"sitehub/service/product/api/internal/svc"
	"sitehub/service/product/api/internal/types"
	"sitehub/service/product/domain"
)

type ChangeDocModeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewChangeDocModeLogic(ctx context.Context, s *svc.ServiceContext) *ChangeDocModeLogic {
	return &ChangeDocModeLogic{ctx, s, logx.WithContext(ctx)}
}

func (l *ChangeDocModeLogic) ChangeDocMode(req *types.ChangeDocModeRequest) (*types.Response, error) {
	_, err := l.svcCtx.Require(l.ctx, "website:docs:edit", "admin", true)
	if err != nil {
		return nil, err
	}
	input := domain.DocModeInput{
		Action:         req.Action,
		ID:             req.ID,
		Label:          req.Label,
		Description:    req.Description,
		Icon:           req.Icon,
		Status:         req.Status,
		DefaultVersion: req.DefaultVersion,
		SortOrder:      int(req.SortOrder),
		Revision:       req.Revision,
	}

	err = l.svcCtx.Core.Content.ChangeDocMode(input)
	return response(struct{}{}, err)
}
