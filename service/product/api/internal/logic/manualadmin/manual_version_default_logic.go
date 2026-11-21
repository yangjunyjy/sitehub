package manualadmin

import (
	"context"
	"sitehub/service/product/api/internal/logic"
	"sitehub/service/product/api/internal/svc"
	"sitehub/service/product/api/internal/types"
)

type ManualVersionDefaultLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewManualVersionDefaultLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ManualVersionDefaultLogic {
	return &ManualVersionDefaultLogic{ctx: ctx, svcCtx: svcCtx}
}
func (l *ManualVersionDefaultLogic) ManualVersionDefault(req *types.ManualNamedRevision) (*types.Response, error) {
	_, err := l.svcCtx.Require(l.ctx, "website:docs:edit", "admin", true)
	if err != nil {
		return nil, err
	}
	return logic.Response(nil, l.svcCtx.Core.Manuals.DefaultVersion(req.ID, req.Revision))
}
