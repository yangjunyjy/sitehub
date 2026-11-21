package logic

import (
	"context"

	"sitehub/service/rbac/rpc/internal/svc"
	"sitehub/service/rbac/rpc/pb/rbac"

	"github.com/zeromicro/go-zero/core/logx"
)

type IconContentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewIconContentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IconContentLogic {
	return &IconContentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *IconContentLogic) IconContent(in *rbac.IconName) (*rbac.IconRecord, error) {
	svg, err := l.svcCtx.Core.Icons.Content(l.ctx, in.Name)
	if err != nil {
		return nil, classify(err)
	}
	return &rbac.IconRecord{Svg: svg}, nil
}
