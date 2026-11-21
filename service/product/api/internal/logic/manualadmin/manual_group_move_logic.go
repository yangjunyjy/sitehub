package manualadmin

import (
	"context"
	"sitehub/service/product/api/internal/logic"
	"sitehub/service/product/api/internal/svc"
	"sitehub/service/product/api/internal/types"
)

type ManualGroupMoveLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewManualGroupMoveLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ManualGroupMoveLogic {
	return &ManualGroupMoveLogic{ctx: ctx, svcCtx: svcCtx}
}
func (l *ManualGroupMoveLogic) ManualGroupMove(req *types.ManualGroupMoveRequest) (*types.Response, error) {
	_, err := l.svcCtx.Require(l.ctx, "website:docs:edit", "admin", true)
	if err != nil {
		return nil, err
	}
	return logic.Response(nil, l.svcCtx.Core.Manuals.MoveGroup(req.Version, req.ID, req.Revision, req.Direction))
}
