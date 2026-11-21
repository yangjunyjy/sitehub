package manualadmin

import (
	"context"
	"sitehub/service/product/api/internal/logic"
	"sitehub/service/product/api/internal/svc"
	"sitehub/service/product/api/internal/types"
	"sitehub/service/product/domain"
)

type ManualGroupSaveLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewManualGroupSaveLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ManualGroupSaveLogic {
	return &ManualGroupSaveLogic{ctx: ctx, svcCtx: svcCtx}
}
func (l *ManualGroupSaveLogic) ManualGroupSave(req *types.ManualGroupRequest) (*types.Response, error) {
	_, err := l.svcCtx.Require(l.ctx, "website:docs:edit", "admin", true)
	if err != nil {
		return nil, err
	}
	return logic.Response(nil, l.svcCtx.Core.Manuals.SaveGroup(domain.DocGroupInput{ID: req.ID, Version: req.Version, Title: req.Title, Revision: req.Revision}))
}
