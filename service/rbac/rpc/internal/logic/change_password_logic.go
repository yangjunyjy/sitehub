package logic

import (
	"context"
	"sitehub/service/rbac/rpc/internal/svc"
	pb "sitehub/service/rbac/rpc/pb/rbac"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ChangePasswordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewChangePasswordLogic(ctx context.Context, s *svc.ServiceContext) *ChangePasswordLogic {
	return &ChangePasswordLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *ChangePasswordLogic) ChangePassword(in *pb.UpdatePasswordRequest) (*pb.Empty, error) {
	actor, err := l.svcCtx.Require(l.ctx, "", "admin", true)
	if err != nil {
		return nil, err
	}
	_ = actor
	if err := validatePassword(in.NewPassword); err != nil {
		return nil, err
	}
	if in.Email != "" && in.Email != actor.Email {
		return nil, status.Error(codes.PermissionDenied, "cannot change another account")
	}
	err = l.svcCtx.Core.Users.UpdatePassword(actor.Email, in.OldPassword, in.NewPassword)
	return done(err)
}
