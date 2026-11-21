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

type WithdrawReleaseFileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWithdrawReleaseFileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WithdrawReleaseFileLogic {
	return &WithdrawReleaseFileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *WithdrawReleaseFileLogic) WithdrawReleaseFile(req *types.ReleaseRevision) (resp *types.Response, err error) {
	if _, err := l.svcCtx.Require(l.ctx, "website:downloads:edit", "admin", true); err != nil {
		return nil, err
	}
	return logic.Response(l.svcCtx.Core.Releases.SetPublished(l.ctx, req.ID, req.Revision, false))
}
