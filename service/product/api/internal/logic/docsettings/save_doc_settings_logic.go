package docsettings

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"sitehub/service/product/api/internal/logic"
	"sitehub/service/product/api/internal/svc"
	"sitehub/service/product/api/internal/types"
	"sitehub/service/product/domain"
)

type SaveDocSettingsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSaveDocSettingsLogic(ctx context.Context, s *svc.ServiceContext) *SaveDocSettingsLogic {
	return &SaveDocSettingsLogic{ctx, s, logx.WithContext(ctx)}
}

func (l *SaveDocSettingsLogic) SaveDocSettings(req *types.SaveDocSettingsRequest) (*types.Response, error) {
	_, err := l.svcCtx.Require(l.ctx, "website:docs:edit", "admin", true)
	if err != nil {
		return nil, err
	}
	input := domain.DocSettingsInput{
		Title:       req.Title,
		Description: req.Description,
		Revision:    req.Revision,
	}

	err = l.svcCtx.Core.Content.SaveDocSettings(input)
	return logic.Response(struct{}{}, err)
}
