package manualadmin

import (
	"context"
	"sitehub/service/product/api/internal/logic"
	"sitehub/service/product/api/internal/svc"
	"sitehub/service/product/api/internal/types"
	"sitehub/service/product/domain"
)

type ManualModeSaveLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewManualModeSaveLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ManualModeSaveLogic {
	return &ManualModeSaveLogic{ctx: ctx, svcCtx: svcCtx}
}
func (l *ManualModeSaveLogic) ManualModeSave(req *types.ManualModeRequest) (*types.Response, error) {
	_, err := l.svcCtx.Require(l.ctx, "website:docs:edit", "admin", true)
	if err != nil {
		return nil, err
	}
	return logic.Response(nil, l.svcCtx.Core.Manuals.SaveMode(domain.DocModeInput{ID: req.ID, Label: req.Label, Description: req.Description, Icon: req.Icon, Status: req.Status, DefaultVersion: req.DefaultVersion, SortOrder: req.SortOrder, Revision: req.Revision}))
}
