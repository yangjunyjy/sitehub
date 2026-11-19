package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"sitehub/service/product/api/internal/svc"
	"sitehub/service/product/api/internal/types"
)

type DocsNavigationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDocsNavigationLogic(ctx context.Context, s *svc.ServiceContext) *DocsNavigationLogic {
	return &DocsNavigationLogic{ctx, s, logx.WithContext(ctx)}
}

func (l *DocsNavigationLogic) DocsNavigation(req *types.DocsNavigationRequest) (*types.Response, error) {
	value, err := l.svcCtx.Core.Content.GetDocsNavigation(req.Version, req.Mode)
	return response(value, err)
}
