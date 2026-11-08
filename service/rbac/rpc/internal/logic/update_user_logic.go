package logic

import (
	"context"
	"sitehub/common/rpcx"
	"sitehub/service/rbac/rpc/internal/svc"
	pb "sitehub/service/rbac/rpc/pb/rbac"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateUserLogic(ctx context.Context, s *svc.ServiceContext) *UpdateUserLogic {
	return &UpdateUserLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *UpdateUserLogic) UpdateUser(in *pb.UpdateSysUserRequest) (*pb.Empty, error) {
	actor, err := l.svcCtx.Require(l.ctx, "system:user:edit", "admin", true)
	if err != nil {
		return nil, err
	}
	_ = actor
	if err := validateUser(in.Username, in.Nickname, in.Email, in.Password, in.Status, in.Gender, false); err != nil {
		return nil, err
	}
	input, err := l.svcCtx.Core.Users.GetUserInfo(in.Id)
	if err != nil {
		return nil, rpcx.Error(err)
	}
	input.Username = in.Username
	input.Nickname = in.Nickname
	input.Email = in.Email
	input.Status = int8(in.Status)
	input.Gender = int8(in.Gender)
	input.Password = in.Password

	err = l.svcCtx.Core.Users.UpdateUser(input, actor.Id)
	return done(err)
}
