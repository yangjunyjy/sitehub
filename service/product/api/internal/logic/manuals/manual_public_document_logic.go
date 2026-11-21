package manuals

import (
	"context"
	"sitehub/service/product/api/internal/logic"
	"sitehub/service/product/api/internal/svc"
	"sitehub/service/product/api/internal/types"
)

type ManualPublicDocumentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewManualPublicDocumentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ManualPublicDocumentLogic {
	return &ManualPublicDocumentLogic{ctx: ctx, svcCtx: svcCtx}
}
func (l *ManualPublicDocumentLogic) ManualPublicDocument(req *types.ManualPublicDocumentQuery) (*types.Response, error) {

	return logic.Response(l.svcCtx.Core.Manuals.PublicDocument(req.ID, req.Version, req.Mode))
}
