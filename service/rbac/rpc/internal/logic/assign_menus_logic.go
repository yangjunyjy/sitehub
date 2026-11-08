package logic

import (
	"context"
	"sitehub/common/protoconv"
	request "sitehub/service/rbac/dto/request"
	"sitehub/service/rbac/rpc/internal/svc"
	pb "sitehub/service/rbac/rpc/pb/rbac"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AssignMenusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAssignMenusLogic(ctx context.Context, s *svc.ServiceContext) *AssignMenusLogic {
	return &AssignMenusLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *AssignMenusLogic) AssignMenus(in *pb.RoleAssignMenusRequest) (*pb.Empty, error) {
	actor, err := l.svcCtx.Require(l.ctx, "system:role:edit", "admin", true)
	if err != nil {
		return nil, err
	}
	_ = actor
	var input request.RoleAssignMenusRequest
	if err := protoconv.Decode(in, &input); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	err = l.svcCtx.Core.Roles.AssignMenus(&input)
	return done(err)
}
