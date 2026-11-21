package downloads

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"sitehub/service/product/api/internal/logic"
	"sitehub/service/product/api/internal/svc"
	"sitehub/service/product/api/internal/types"
)

type DownloadsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDownloadsLogic(ctx context.Context, s *svc.ServiceContext) *DownloadsLogic {
	return &DownloadsLogic{ctx, s, logx.WithContext(ctx)}
}

func (l *DownloadsLogic) Downloads() (*types.Response, error) {
	value, err := l.svcCtx.Core.Content.GetPublicDownloads()
	return logic.Response(value, err)
}
