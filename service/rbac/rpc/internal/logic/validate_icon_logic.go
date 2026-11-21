package logic

import (
	"context"
	services "sitehub/service/rbac/business"

	"sitehub/service/rbac/rpc/internal/svc"
	"sitehub/service/rbac/rpc/pb/rbac"

	"github.com/zeromicro/go-zero/core/logx"
)

type ValidateIconLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewValidateIconLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ValidateIconLogic {
	return &ValidateIconLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ValidateIconLogic) ValidateIcon(in *rbac.IconRecord) (*rbac.IconRecord, error) {
	if _, err := l.svcCtx.Require(l.ctx, "system:icon:create", "admin", false); err != nil {
		return nil, err
	}
	input := services.SystemIconInput{Name: in.Name, Label: in.Label, Category: in.Category, SVG: in.Svg}
	icon, err := l.svcCtx.Core.Icons.Validate(input)
	return reply[rbac.IconRecord](icon, err)
}
