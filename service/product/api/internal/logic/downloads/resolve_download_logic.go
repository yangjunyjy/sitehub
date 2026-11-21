package downloads

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"sitehub/service/product/api/internal/logic"
	"sitehub/service/product/api/internal/svc"
	"sitehub/service/product/api/internal/types"
)

type ResolveDownloadLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewResolveDownloadLogic(ctx context.Context, s *svc.ServiceContext) *ResolveDownloadLogic {
	return &ResolveDownloadLogic{ctx, s, logx.WithContext(ctx)}
}

func (l *ResolveDownloadLogic) ResolveDownload(req *types.ResolveDownloadRequest) (*types.Response, error) {
	_, err := l.svcCtx.Require(l.ctx, "", "member", false)
	if err != nil {
		return nil, err
	}
	value, err := l.svcCtx.Core.Content.GetPublishedDownload(req.ID)
	return logic.Response(value, err)
}
