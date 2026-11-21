package manualadmin

import (
	"context"
	"sitehub/service/product/api/internal/logic"
	"sitehub/service/product/api/internal/svc"
	"sitehub/service/product/api/internal/types"
)

type ManualModeDeleteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewManualModeDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ManualModeDeleteLogic {
	return &ManualModeDeleteLogic{ctx: ctx, svcCtx: svcCtx}
}
func (l *ManualModeDeleteLogic) ManualModeDelete(req *types.ManualNamedRevision) (*types.Response, error) {
	_, err := l.svcCtx.Require(l.ctx, "website:docs:edit", "admin", true)
	if err != nil {
		return nil, err
	}
	return logic.Response(nil, l.svcCtx.Core.Manuals.DeleteMode(req.ID, req.Revision))
}
