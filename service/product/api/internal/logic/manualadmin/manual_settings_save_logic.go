package manualadmin

import (
	"context"
	"sitehub/service/product/api/internal/logic"
	"sitehub/service/product/api/internal/svc"
	"sitehub/service/product/api/internal/types"
	"sitehub/service/product/domain"
)

type ManualSettingsSaveLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewManualSettingsSaveLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ManualSettingsSaveLogic {
	return &ManualSettingsSaveLogic{ctx: ctx, svcCtx: svcCtx}
}
func (l *ManualSettingsSaveLogic) ManualSettingsSave(req *types.ManualSettingsRequest) (*types.Response, error) {
	_, err := l.svcCtx.Require(l.ctx, "website:docs:edit", "admin", true)
	if err != nil {
		return nil, err
	}
	return logic.Response(nil, l.svcCtx.Core.Manuals.SaveDocSettings(domain.DocSettingsInput{Title: req.Title, Description: req.Description, Revision: req.Revision}))
}
