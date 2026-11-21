package auth

import (
	"context"

	"sitehub/common/protoconv"
	"sitehub/common/rpcx"
	"sitehub/service/rbac/api/internal/session"
	"sitehub/service/rbac/api/internal/svc"
	"sitehub/service/rbac/api/internal/types"
	pb "sitehub/service/rbac/rpc/pb/rbac"
)

func Logout(ctx context.Context, services *svc.ServiceContext) (*types.Response, error) {
	in := &pb.Empty{}
	out, err := services.RPC.Logout(rpcx.Outgoing(ctx, services.Config.ServiceKey), in)
	if err != nil {
		return nil, err
	}
	value, err := protoconv.Value(out)
	if err != nil {
		return nil, err
	}
	session.Clear(ctx, "admin")
	return &types.Response{Code: 0, Message: "ok", Data: value}, nil
}
