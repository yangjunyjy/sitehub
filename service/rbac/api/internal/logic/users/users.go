package users

import (
	"context"

	"sitehub/service/rbac/api/internal/logic"
	"sitehub/service/rbac/api/internal/svc"
	"sitehub/service/rbac/api/internal/types"
	pb "sitehub/service/rbac/rpc/pb/rbac"
)

func ListUsers(ctx context.Context, services *svc.ServiceContext, req *types.ListUsersRequest) (*types.Response, error) {
	return logic.Forward(ctx, services, req, &pb.SearchSysUserQueryRequest{}, services.RPC.ListUsers)
}

func GetUser(ctx context.Context, services *svc.ServiceContext, req *types.GetUserRequest) (*types.Response, error) {
	return logic.Forward(ctx, services, req, &pb.IDRequest{}, services.RPC.GetUser)
}

func CreateUser(ctx context.Context, services *svc.ServiceContext, req *types.CreateUserRequest) (*types.Response, error) {
	return logic.Forward(ctx, services, req, &pb.CreateSysUserRequest{}, services.RPC.CreateUser)
}

func UpdateUser(ctx context.Context, services *svc.ServiceContext, req *types.UpdateUserRequest) (*types.Response, error) {
	return logic.Forward(ctx, services, req, &pb.UpdateSysUserRequest{}, services.RPC.UpdateUser)
}

func DeleteUser(ctx context.Context, services *svc.ServiceContext, req *types.DeleteUserRequest) (*types.Response, error) {
	return logic.Forward(ctx, services, req, &pb.IDRequest{}, services.RPC.DeleteUser)
}

func AssignRoles(ctx context.Context, services *svc.ServiceContext, req *types.AssignRolesRequest) (*types.Response, error) {
	return logic.Forward(ctx, services, req, &pb.AssignRolesRequest{}, services.RPC.AssignRoles)
}
