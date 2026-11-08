package logic

import (
	"context"
	"sitehub/common/protoconv"
	request "sitehub/service/rbac/dto/request"
	"sitehub/service/rbac/rpc/internal/svc"
	pb "sitehub/service/rbac/rpc/pb/rbac"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListRolesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListRolesLogic(ctx context.Context, s *svc.ServiceContext) *ListRolesLogic {
	return &ListRolesLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *ListRolesLogic) ListRoles(in *pb.RoleSearchQueryRequest) (*pb.RolePage, error) {
	actor, err := l.svcCtx.Require(l.ctx, "system:role:list", "admin", false)
	if err != nil {
		return nil, err
	}
	_ = actor
	var input request.RoleSearchQueryRequest
	if err := protoconv.Decode(in, &input); err != nil {
		return nil, err
	}
	if input.Page == 0 {
		input.Page = 1
	}
	if input.PageSize == 0 {
		input.PageSize = 20
	}
	items, total, err := l.svcCtx.Core.Roles.List(&input)
	return reply[pb.RolePage](map[string]any{"items": items, "total": total}, err)
}
