package roles

import (
	"context"

	"sitehub/service/rbac/api/internal/logic"
	"sitehub/service/rbac/api/internal/svc"
	"sitehub/service/rbac/api/internal/types"
	pb "sitehub/service/rbac/rpc/pb/rbac"
)

func ListRoles(ctx context.Context, services *svc.ServiceContext, req *types.ListRolesRequest) (*types.Response, error) {
	return logic.Forward(ctx, services, req, &pb.RoleSearchQueryRequest{}, services.RPC.ListRoles)
}

func GetRole(ctx context.Context, services *svc.ServiceContext, req *types.GetRoleRequest) (*types.Response, error) {
	return logic.Forward(ctx, services, req, &pb.IDRequest{}, services.RPC.GetRole)
}

func CreateRole(ctx context.Context, services *svc.ServiceContext, req *types.CreateRoleRequest) (*types.Response, error) {
	return logic.Forward(ctx, services, req, &pb.CreateRoleRequest{}, services.RPC.CreateRole)
}

func UpdateRole(ctx context.Context, services *svc.ServiceContext, req *types.UpdateRoleRequest) (*types.Response, error) {
	return logic.Forward(ctx, services, req, &pb.UpdateRoleRequest{}, services.RPC.UpdateRole)
}

func DeleteRole(ctx context.Context, services *svc.ServiceContext, req *types.DeleteRoleRequest) (*types.Response, error) {
	return logic.Forward(ctx, services, req, &pb.IDRequest{}, services.RPC.DeleteRole)
}

func AssignMenus(ctx context.Context, services *svc.ServiceContext, req *types.AssignMenusRequest) (*types.Response, error) {
	return logic.Forward(ctx, services, req, &pb.RoleAssignMenusRequest{}, services.RPC.AssignMenus)
}

func RoleMenus(ctx context.Context, services *svc.ServiceContext, req *types.RoleMenusRequest) (*types.Response, error) {
	return logic.Forward(ctx, services, req, &pb.IDRequest{}, services.RPC.RoleMenus)
}
