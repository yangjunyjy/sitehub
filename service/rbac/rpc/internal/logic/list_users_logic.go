package logic

import (
	"context"
	"sitehub/common/protoconv"
	request "sitehub/service/rbac/dto/request"
	"sitehub/service/rbac/rpc/internal/svc"
	pb "sitehub/service/rbac/rpc/pb/rbac"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListUsersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListUsersLogic(ctx context.Context, s *svc.ServiceContext) *ListUsersLogic {
	return &ListUsersLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *ListUsersLogic) ListUsers(in *pb.SearchSysUserQueryRequest) (*pb.UserPage, error) {
	actor, err := l.svcCtx.Require(l.ctx, "system:user:list", "admin", false)
	if err != nil {
		return nil, err
	}
	_ = actor
	var input request.SearchSysUserQueryRequest
	if err := protoconv.Decode(in, &input); err != nil {
		return nil, err
	}
	if input.Page == 0 {
		input.Page = 1
	}
	if input.PageSize == 0 {
		input.PageSize = 20
	}
	items, total, err := l.svcCtx.Core.Users.List(&input)
	return reply[pb.UserPage](map[string]any{"items": items, "total": total}, err)
}
