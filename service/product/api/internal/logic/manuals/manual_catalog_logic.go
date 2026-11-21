package manuals

import (
	"context"
	"sitehub/service/product/api/internal/logic"
	"sitehub/service/product/api/internal/svc"
	"sitehub/service/product/api/internal/types"
)

type ManualCatalogLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewManualCatalogLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ManualCatalogLogic {
	return &ManualCatalogLogic{ctx: ctx, svcCtx: svcCtx}
}
func (l *ManualCatalogLogic) ManualCatalog(req *types.ManualCatalogQuery) (*types.Response, error) {

	return logic.Response(l.svcCtx.Core.Manuals.GetDocCatalog(req.Version, req.Mode))
}
