// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package releaseadmin

import (
	"context"
	"sitehub/service/product/api/internal/logic"

	"sitehub/service/product/api/internal/svc"
	"sitehub/service/product/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ReleaseFilesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReleaseFilesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReleaseFilesLogic {
	return &ReleaseFilesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ReleaseFilesLogic) ReleaseFiles() (resp *types.Response, err error) {
	if _, err := l.svcCtx.Require(l.ctx, "website:downloads:edit", "admin", false); err != nil {
		return nil, err
	}
	items, e := l.svcCtx.Core.Releases.Files(l.ctx)
	return logic.Response(map[string]any{"items": items}, e)
}
