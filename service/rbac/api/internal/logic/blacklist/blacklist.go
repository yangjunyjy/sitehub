package blacklist

import (
	"context"

	"sitehub/service/rbac/api/internal/logic"
	"sitehub/service/rbac/api/internal/svc"
	"sitehub/service/rbac/api/internal/types"
	pb "sitehub/service/rbac/rpc/pb/rbac"
)

func Blacklist(ctx context.Context, services *svc.ServiceContext) (*types.Response, error) {
	return logic.Forward(ctx, services, nil, &pb.Empty{}, services.RPC.Blacklist)
}

func BlockIP(ctx context.Context, services *svc.ServiceContext, req *types.BlockIPRequest) (*types.Response, error) {
	return logic.Forward(ctx, services, req, &pb.BlacklistInput{}, services.RPC.BlockIP)
}

func UnblockIP(ctx context.Context, services *svc.ServiceContext, req *types.UnblockIPRequest) (*types.Response, error) {
	return logic.Forward(ctx, services, req, &pb.BlacklistInput{}, services.RPC.UnblockIP)
}
