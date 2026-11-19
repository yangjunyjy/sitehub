package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"sitehub/service/product/api/internal/svc"
	"sitehub/service/product/api/internal/types"
)

type DocsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDocsLogic(ctx context.Context, s *svc.ServiceContext) *DocsLogic {
	return &DocsLogic{ctx, s, logx.WithContext(ctx)}
}

func (l *DocsLogic) Docs() (*types.Response, error) {
	_, err := l.svcCtx.Require(l.ctx, "website:docs:edit", "admin", false)
	if err != nil {
		return nil, err
	}
	value, err := l.svcCtx.Core.Content.GetDocs()
	return response(value, err)
}
