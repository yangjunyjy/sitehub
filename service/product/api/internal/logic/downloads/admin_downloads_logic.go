package downloads

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"sitehub/service/product/api/internal/logic"
	"sitehub/service/product/api/internal/svc"
	"sitehub/service/product/api/internal/types"
)

type AdminDownloadsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminDownloadsLogic(ctx context.Context, s *svc.ServiceContext) *AdminDownloadsLogic {
	return &AdminDownloadsLogic{ctx, s, logx.WithContext(ctx)}
}

func (l *AdminDownloadsLogic) AdminDownloads() (*types.Response, error) {
	_, err := l.svcCtx.Require(l.ctx, "website:downloads:edit", "admin", false)
	if err != nil {
		return nil, err
	}
	value, err := l.svcCtx.Core.Content.GetDownloads()
	return logic.Response(value, err)
}
