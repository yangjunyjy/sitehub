package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"sitehub/service/product/api/internal/svc"
	"sitehub/service/product/api/internal/types"
)

type DocModesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDocModesLogic(ctx context.Context, s *svc.ServiceContext) *DocModesLogic {
	return &DocModesLogic{ctx, s, logx.WithContext(ctx)}
}

func (l *DocModesLogic) DocModes() (*types.Response, error) {
	value, err := l.svcCtx.Core.Content.GetDocModes()
	return response(value, err)
}
