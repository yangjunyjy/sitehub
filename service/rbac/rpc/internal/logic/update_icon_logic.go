package logic

import (
	"context"
	services "sitehub/service/rbac/business"

	"sitehub/service/rbac/rpc/internal/svc"
	"sitehub/service/rbac/rpc/pb/rbac"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateIconLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateIconLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateIconLogic {
	return &UpdateIconLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateIconLogic) UpdateIcon(in *rbac.IconRecord) (*rbac.IconRecord, error) {
	if _, err := l.svcCtx.Require(l.ctx, "system:icon:update", "admin", true); err != nil {
		return nil, err
	}
	input := services.SystemIconInput{Name: in.Name, Label: in.Label, Category: in.Category, SVG: in.Svg}
	icon, err := l.svcCtx.Core.Icons.Update(l.ctx, in.Id, input)
	return reply[rbac.IconRecord](icon, err)
}
