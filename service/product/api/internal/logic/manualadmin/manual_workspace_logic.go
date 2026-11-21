package manualadmin

import (
	"context"
	"sitehub/service/product/api/internal/logic"
	"sitehub/service/product/api/internal/svc"
	"sitehub/service/product/api/internal/types"
)

type ManualWorkspaceLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewManualWorkspaceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ManualWorkspaceLogic {
	return &ManualWorkspaceLogic{ctx: ctx, svcCtx: svcCtx}
}
func (l *ManualWorkspaceLogic) ManualWorkspace(req *types.ManualCatalogQuery) (*types.Response, error) {
	_, err := l.svcCtx.Require(l.ctx, "website:docs:edit", "admin", false)
	if err != nil {
		return nil, err
	}
	return logic.Response(l.svcCtx.Core.Manuals.Workspace(req.Mode, req.Version))
}
