package manualadmin

import (
	"context"
	"sitehub/service/product/api/internal/logic"
	"sitehub/service/product/api/internal/svc"
	"sitehub/service/product/api/internal/types"
	"sitehub/service/product/domain"
)

type ManualVersionSaveLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewManualVersionSaveLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ManualVersionSaveLogic {
	return &ManualVersionSaveLogic{ctx: ctx, svcCtx: svcCtx}
}
func (l *ManualVersionSaveLogic) ManualVersionSave(req *types.ManualVersionRequest) (*types.Response, error) {
	actor, err := l.svcCtx.Require(l.ctx, "website:docs:edit", "admin", true)
	if err != nil {
		return nil, err
	}
	return logic.Response(nil, l.svcCtx.Core.Manuals.SaveVersion(domain.DocVersionInput{ID: req.ID, Name: req.Name, ModeID: req.ModeID, Source: req.Source, Revision: req.Revision}, actor.Id))
}
