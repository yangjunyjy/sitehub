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

type UpdateRoleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateRoleLogic(ctx context.Context, s *svc.ServiceContext) *UpdateRoleLogic {
	return &UpdateRoleLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *UpdateRoleLogic) UpdateRole(in *pb.UpdateRoleRequest) (*pb.Empty, error) {
	actor, err := l.svcCtx.Require(l.ctx, "system:role:edit", "admin", true)
	if err != nil {
		return nil, err
	}
	_ = actor
	if err := validateRole(in.RoleCode, in.RoleName, in.RoleDesc, in.Status); err != nil {
		return nil, err
	}
	var input entity.SysRole
	if err := protoconv.Decode(in, &input); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	err = l.svcCtx.Core.Roles.Update(&input)
	return done(err)
}
