package manualadmin

import (
	"context"
	"sitehub/service/product/api/internal/logic"
	"sitehub/service/product/api/internal/svc"
	"sitehub/service/product/api/internal/types"
)

type ManualHistoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewManualHistoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ManualHistoryLogic {
	return &ManualHistoryLogic{ctx: ctx, svcCtx: svcCtx}
}
func (l *ManualHistoryLogic) ManualHistory(req *types.ManualDocumentQuery) (*types.Response, error) {
	_, err := l.svcCtx.Require(l.ctx, "website:docs:edit", "admin", false)
	if err != nil {
		return nil, err
	}
	return logic.Response(l.svcCtx.Core.Manuals.History(req.Version, req.ID))
}
