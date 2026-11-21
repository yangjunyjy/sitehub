package manualadmin

import (
	"context"
	"sitehub/service/product/api/internal/logic"
	"sitehub/service/product/api/internal/svc"
	"sitehub/service/product/api/internal/types"
)

type ManualModeDefaultLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewManualModeDefaultLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ManualModeDefaultLogic {
	return &ManualModeDefaultLogic{ctx: ctx, svcCtx: svcCtx}
}
func (l *ManualModeDefaultLogic) ManualModeDefault(req *types.ManualNamedRevision) (*types.Response, error) {
	_, err := l.svcCtx.Require(l.ctx, "website:docs:edit", "admin", true)
	if err != nil {
		return nil, err
	}
	return logic.Response(nil, l.svcCtx.Core.Manuals.DefaultMode(req.ID, req.Revision))
}
