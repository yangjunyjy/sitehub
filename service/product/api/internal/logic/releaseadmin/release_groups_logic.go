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

type ReleaseGroupsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReleaseGroupsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReleaseGroupsLogic {
	return &ReleaseGroupsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ReleaseGroupsLogic) ReleaseGroups() (resp *types.Response, err error) {
	if _, err := l.svcCtx.Require(l.ctx, "website:downloads:edit", "admin", false); err != nil {
		return nil, err
	}
	items, e := l.svcCtx.Core.Releases.Groups(l.ctx)
	return logic.Response(map[string]any{"items": items}, e)
}
