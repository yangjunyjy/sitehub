package logic

import (
	"context"
	"sitehub/common/protoconv"
	entity "sitehub/service/rbac/model/entity"
	"sitehub/service/rbac/rpc/internal/svc"
	pb "sitehub/service/rbac/rpc/pb/rbac"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CreateUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateUserLogic(ctx context.Context, s *svc.ServiceContext) *CreateUserLogic {
	return &CreateUserLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *CreateUserLogic) CreateUser(in *pb.CreateSysUserRequest) (*pb.Empty, error) {
	actor, err := l.svcCtx.Require(l.ctx, "system:user:add", "admin", true)
	if err != nil {
		return nil, err
	}
	_ = actor
	if err := validateUser(in.Username, in.Nickname, in.Email, in.Password, in.Status, in.Gender, true); err != nil {
		return nil, err
	}
	var input entity.SysUser
	if err := protoconv.Decode(in, &input); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	input.Password = in.Password
	err = l.svcCtx.Core.Users.CreateUser(&input, actor.Id)
	return done(err)
}
