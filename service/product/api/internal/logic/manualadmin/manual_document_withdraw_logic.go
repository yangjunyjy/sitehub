package manualadmin

import (
	"context"
	"sitehub/service/product/api/internal/logic"
	"sitehub/service/product/api/internal/svc"
	"sitehub/service/product/api/internal/types"
	"sitehub/service/product/domain"
)

type ManualDocumentWithdrawLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewManualDocumentWithdrawLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ManualDocumentWithdrawLogic {
	return &ManualDocumentWithdrawLogic{ctx: ctx, svcCtx: svcCtx}
}
func (l *ManualDocumentWithdrawLogic) ManualDocumentWithdraw(req *types.ManualDocumentRevision) (*types.Response, error) {
	actor, err := l.svcCtx.Require(l.ctx, "website:docs:edit", "admin", true)
	if err != nil {
		return nil, err
	}
	return logic.Response(l.svcCtx.Core.Manuals.SetDocumentPublished(domain.DocumentDeleteInput{ID: req.ID, Version: req.Version, Revision: req.Revision}, actor.Id, false))
}
