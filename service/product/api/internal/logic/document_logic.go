package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"sitehub/service/product/api/internal/svc"
	"sitehub/service/product/api/internal/types"
)

type DocumentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDocumentLogic(ctx context.Context, s *svc.ServiceContext) *DocumentLogic {
	return &DocumentLogic{ctx, s, logx.WithContext(ctx)}
}

func (l *DocumentLogic) Document(req *types.DocumentRequest) (*types.Response, error) {
	value, err := l.svcCtx.Core.Content.GetDocument(req.ID, req.Version, req.Mode)
	return response(value, err)
}
