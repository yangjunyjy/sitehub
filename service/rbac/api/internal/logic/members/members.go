package members

import (
	"context"

	"sitehub/service/rbac/api/internal/logic"
	"sitehub/service/rbac/api/internal/svc"
	"sitehub/service/rbac/api/internal/types"
	pb "sitehub/service/rbac/rpc/pb/rbac"
)

func SendMemberCode(ctx context.Context, services *svc.ServiceContext, req *types.SendMemberCodeRequest) (*types.Response, error) {
	return logic.Forward(ctx, services, req, &pb.SendCodeRequest{}, services.RPC.SendMemberCode)
}

func RegisterMember(ctx context.Context, services *svc.ServiceContext, req *types.RegisterMemberRequest) (*types.Response, error) {
	return logic.Forward(ctx, services, req, &pb.RegisterMemberInput{}, services.RPC.RegisterMember)
}

func CurrentMember(ctx context.Context, services *svc.ServiceContext) (*types.Response, error) {
	return logic.Forward(ctx, services, nil, &pb.Empty{}, services.RPC.CurrentMember)
}
