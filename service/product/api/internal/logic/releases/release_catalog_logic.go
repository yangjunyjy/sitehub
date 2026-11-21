// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package releases

import (
	"context"
	"sitehub/service/product/api/internal/logic"

	"sitehub/service/product/api/internal/svc"
	"sitehub/service/product/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ReleaseCatalogLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReleaseCatalogLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReleaseCatalogLogic {
	return &ReleaseCatalogLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ReleaseCatalogLogic) ReleaseCatalog() (resp *types.Response, err error) {
	return logic.Response(l.svcCtx.Core.Releases.Catalog(l.ctx))
}
