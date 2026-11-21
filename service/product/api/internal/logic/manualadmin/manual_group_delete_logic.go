package manualadmin

import (
	"context"
	"sitehub/service/product/api/internal/logic"
	"sitehub/service/product/api/internal/svc"
	"sitehub/service/product/api/internal/types"
)

type ManualGroupDeleteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewManualGroupDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ManualGroupDeleteLogic {
	return &ManualGroupDeleteLogic{ctx: ctx, svcCtx: svcCtx}
}
func (l *ManualGroupDeleteLogic) ManualGroupDelete(req *types.ManualGroupRevision) (*types.Response, error) {
	_, err := l.svcCtx.Require(l.ctx, "website:docs:edit", "admin", true)
	if err != nil {
		return nil, err
	}
	return logic.Response(nil, l.svcCtx.Core.Manuals.DeleteGroup(req.Version, req.ID, req.Revision))
}
