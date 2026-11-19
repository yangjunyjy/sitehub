package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"sitehub/service/product/api/internal/svc"
	"sitehub/service/product/api/internal/types"
)

type DocumentHistoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDocumentHistoryLogic(ctx context.Context, s *svc.ServiceContext) *DocumentHistoryLogic {
	return &DocumentHistoryLogic{ctx, s, logx.WithContext(ctx)}
}

func (l *DocumentHistoryLogic) DocumentHistory(req *types.DocumentHistoryRequest) (*types.Response, error) {
	_, err := l.svcCtx.Require(l.ctx, "website:docs:edit", "admin", false)
	if err != nil {
		return nil, err
	}
	value, err := l.svcCtx.Core.Content.DocumentHistory(req.ID, req.Version, req.HistoryID)
	if req.HistoryID > 0 {
		return response(map[string]any{"detail": value}, err)
	}
	return response(map[string]any{"items": value}, err)
}
