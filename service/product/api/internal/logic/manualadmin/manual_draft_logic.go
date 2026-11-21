package manualadmin

import (
	"context"
	"sitehub/service/product/api/internal/logic"
	"sitehub/service/product/api/internal/svc"
	"sitehub/service/product/api/internal/types"
)

type ManualDraftLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewManualDraftLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ManualDraftLogic {
	return &ManualDraftLogic{ctx: ctx, svcCtx: svcCtx}
}
func (l *ManualDraftLogic) ManualDraft(req *types.ManualDocumentQuery) (*types.Response, error) {
	_, err := l.svcCtx.Require(l.ctx, "website:docs:edit", "admin", false)
	if err != nil {
		return nil, err
	}
	return logic.Response(l.svcCtx.Core.Manuals.Draft(req.ID, req.Version))
}
