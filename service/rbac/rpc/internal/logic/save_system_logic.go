package logic

import (
	"context"
	"sitehub/common/protoconv"
	"sitehub/service/rbac/model/entity"

	"sitehub/service/rbac/rpc/internal/svc"
	"sitehub/service/rbac/rpc/pb/rbac"

	"github.com/zeromicro/go-zero/core/logx"
)

type SaveSystemLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSaveSystemLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveSystemLogic {
	return &SaveSystemLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SaveSystemLogic) SaveSystem(in *rbac.SystemEntry) (*rbac.Empty, error) {
	permission := "system:system:create"
	if in.Id != 0 {
		permission = "system:system:update"
	}
	if _, err := l.svcCtx.Require(l.ctx, permission, "admin", true); err != nil {
		return nil, err
	}
	var value entity.SysSystem
	if err := protoconv.Decode(in, &value); err != nil {
		return done(err)
	}
	return done(l.svcCtx.Core.Systems.Save(l.ctx, &value))
}
