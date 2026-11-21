package logic

import (
	"context"
	services "sitehub/service/rbac/business"

	"sitehub/service/rbac/rpc/internal/svc"
	"sitehub/service/rbac/rpc/pb/rbac"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateIconLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateIconLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateIconLogic {
	return &CreateIconLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateIconLogic) CreateIcon(in *rbac.IconRecord) (*rbac.IconRecord, error) {
	if _, err := l.svcCtx.Require(l.ctx, "system:icon:create", "admin", true); err != nil {
		return nil, err
	}
	input := services.SystemIconInput{Name: in.Name, Label: in.Label, Category: in.Category, SVG: in.Svg}
	icon, err := l.svcCtx.Core.Icons.Create(l.ctx, input)
	return reply[rbac.IconRecord](icon, err)
}
