package documents

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"sitehub/service/product/api/internal/logic"
	"sitehub/service/product/api/internal/svc"
	"sitehub/service/product/api/internal/types"
)

type DocCatalogLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDocCatalogLogic(ctx context.Context, s *svc.ServiceContext) *DocCatalogLogic {
	return &DocCatalogLogic{ctx, s, logx.WithContext(ctx)}
}

func (l *DocCatalogLogic) DocCatalog(req *types.DocCatalogRequest) (*types.Response, error) {
	value, err := l.svcCtx.Core.Content.GetDocCatalog(req.Version, req.Mode)
	return logic.Response(value, err)
}
