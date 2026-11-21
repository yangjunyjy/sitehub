package menus

import (
	"context"

	"sitehub/service/rbac/api/internal/logic"
	"sitehub/service/rbac/api/internal/svc"
	"sitehub/service/rbac/api/internal/types"
	pb "sitehub/service/rbac/rpc/pb/rbac"
)

func MenuList(ctx context.Context, services *svc.ServiceContext, req *types.SystemRequest) (*types.Response, error) {
	return MenuTree(ctx, services, req)
}

func MenuTree(ctx context.Context, services *svc.ServiceContext, req *types.SystemRequest) (*types.Response, error) {
	return logic.Forward(ctx, services, nil, &pb.SystemRequest{System: req.System}, services.RPC.MenuTree)
}

func PublicMenus(ctx context.Context, services *svc.ServiceContext, req *types.SystemRequest) (*types.Response, error) {
	return logic.Forward(ctx, services, req, &pb.SystemRequest{}, services.RPC.PublicMenus)
}

func GetMenu(ctx context.Context, services *svc.ServiceContext, req *types.GetMenuRequest) (*types.Response, error) {
	return logic.Forward(ctx, services, req, &pb.IDRequest{}, services.RPC.GetMenu)
}

func CreateMenu(ctx context.Context, services *svc.ServiceContext, req *types.CreateMenuRequest) (*types.Response, error) {
	return logic.Forward(ctx, services, req, &pb.CreateSysMenuRequest{}, services.RPC.CreateMenu)
}

func UpdateMenu(ctx context.Context, services *svc.ServiceContext, req *types.UpdateMenuRequest) (*types.Response, error) {
	return logic.Forward(ctx, services, req, &pb.UpdateSysMenuRequest{}, services.RPC.UpdateMenu)
}

func DeleteMenu(ctx context.Context, services *svc.ServiceContext, req *types.DeleteMenuRequest) (*types.Response, error) {
	return logic.Forward(ctx, services, req, &pb.IDRequest{}, services.RPC.DeleteMenu)
}
