package manualadmin

import (
	"context"
	"sitehub/service/product/api/internal/logic"
	"sitehub/service/product/api/internal/svc"
	"sitehub/service/product/api/internal/types"
)

type ManualVersionPublishLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewManualVersionPublishLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ManualVersionPublishLogic {
	return &ManualVersionPublishLogic{ctx: ctx, svcCtx: svcCtx}
}
func (l *ManualVersionPublishLogic) ManualVersionPublish(req *types.ManualNamedRevision) (*types.Response, error) {
	_, err := l.svcCtx.Require(l.ctx, "website:docs:edit", "admin", true)
	if err != nil {
		return nil, err
	}
	return logic.Response(nil, l.svcCtx.Core.Manuals.SetVersionPublished(req.ID, req.Revision, true))
}
