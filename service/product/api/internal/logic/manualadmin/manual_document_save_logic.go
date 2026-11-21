package manualadmin

import (
	"context"
	"sitehub/service/product/api/internal/logic"
	"sitehub/service/product/api/internal/svc"
	"sitehub/service/product/api/internal/types"
	"sitehub/service/product/domain"
)

type ManualDocumentSaveLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewManualDocumentSaveLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ManualDocumentSaveLogic {
	return &ManualDocumentSaveLogic{ctx: ctx, svcCtx: svcCtx}
}
func (l *ManualDocumentSaveLogic) ManualDocumentSave(req *types.ManualDraftRequest) (*types.Response, error) {
	actor, err := l.svcCtx.Require(l.ctx, "website:docs:edit", "admin", true)
	if err != nil {
		return nil, err
	}
	return logic.Response(l.svcCtx.Core.Manuals.SaveDocument(domain.DocumentInput{ID: req.ID, Version: req.Version, Slug: req.Slug, GroupID: req.GroupID, Title: req.Title, Content: req.Content, Summary: req.Summary, Revision: req.Revision}, actor.Id))
}
