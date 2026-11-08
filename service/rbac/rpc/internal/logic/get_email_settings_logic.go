package logic

import (
	"context"
	"sitehub/service/rbac/rpc/internal/svc"
	pb "sitehub/service/rbac/rpc/pb/rbac"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetEmailSettingsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetEmailSettingsLogic(ctx context.Context, s *svc.ServiceContext) *GetEmailSettingsLogic {
	return &GetEmailSettingsLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *GetEmailSettingsLogic) GetEmailSettings(in *pb.Empty) (*pb.EmailSettingsView, error) {
	actor, err := l.svcCtx.Require(l.ctx, "system:config:list", "admin", false)
	if err != nil {
		return nil, err
	}
	_ = actor
	value, err := l.svcCtx.Core.Email.Get()
	return reply[pb.EmailSettingsView](value, err)
}
