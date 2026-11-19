package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"sitehub/service/product/api/internal/svc"
	"sitehub/service/product/api/internal/types"
)

type AdminDocumentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminDocumentLogic(ctx context.Context, s *svc.ServiceContext) *AdminDocumentLogic {
	return &AdminDocumentLogic{ctx, s, logx.WithContext(ctx)}
}

func (l *AdminDocumentLogic) AdminDocument(req *types.AdminDocumentRequest) (*types.Response, error) {
	_, err := l.svcCtx.Require(l.ctx, "website:docs:edit", "admin", false)
	if err != nil {
		return nil, err
	}
	value, err := l.svcCtx.Core.Content.GetAdminDocument(req.ID, req.Version)
	return response(value, err)
}
