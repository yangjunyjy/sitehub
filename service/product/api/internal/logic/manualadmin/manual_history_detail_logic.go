package manualadmin

import (
	"context"
	"sitehub/service/product/api/internal/logic"
	"sitehub/service/product/api/internal/svc"
	"sitehub/service/product/api/internal/types"
)

type ManualHistoryDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewManualHistoryDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ManualHistoryDetailLogic {
	return &ManualHistoryDetailLogic{ctx: ctx, svcCtx: svcCtx}
}
func (l *ManualHistoryDetailLogic) ManualHistoryDetail(req *types.ManualHistoryQuery) (*types.Response, error) {
	_, err := l.svcCtx.Require(l.ctx, "website:docs:edit", "admin", false)
	if err != nil {
		return nil, err
	}
	return logic.Response(l.svcCtx.Core.Manuals.Historical(req.Version, req.ID, req.HistoryID))
}
