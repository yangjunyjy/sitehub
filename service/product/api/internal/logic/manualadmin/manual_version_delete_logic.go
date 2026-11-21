package manualadmin

import (
	"context"
	"sitehub/service/product/api/internal/logic"
	"sitehub/service/product/api/internal/svc"
	"sitehub/service/product/api/internal/types"
)

type ManualVersionDeleteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewManualVersionDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ManualVersionDeleteLogic {
	return &ManualVersionDeleteLogic{ctx: ctx, svcCtx: svcCtx}
}
func (l *ManualVersionDeleteLogic) ManualVersionDelete(req *types.ManualNamedRevision) (*types.Response, error) {
	_, err := l.svcCtx.Require(l.ctx, "website:docs:edit", "admin", true)
	if err != nil {
		return nil, err
	}
	return logic.Response(nil, l.svcCtx.Core.Manuals.DeleteVersion(req.ID, req.Revision))
}
