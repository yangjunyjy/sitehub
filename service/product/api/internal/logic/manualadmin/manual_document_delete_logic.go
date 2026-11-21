package manualadmin

import (
	"context"
	"sitehub/service/product/api/internal/logic"
	"sitehub/service/product/api/internal/svc"
	"sitehub/service/product/api/internal/types"
	"sitehub/service/product/domain"
)

type ManualDocumentDeleteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewManualDocumentDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ManualDocumentDeleteLogic {
	return &ManualDocumentDeleteLogic{ctx: ctx, svcCtx: svcCtx}
}
func (l *ManualDocumentDeleteLogic) ManualDocumentDelete(req *types.ManualDocumentRevision) (*types.Response, error) {
	actor, err := l.svcCtx.Require(l.ctx, "website:docs:edit", "admin", true)
	if err != nil {
		return nil, err
	}
	return logic.Response(nil, l.svcCtx.Core.Manuals.DeleteDocument(domain.DocumentDeleteInput{ID: req.ID, Version: req.Version, Revision: req.Revision}, actor.Id))
}
