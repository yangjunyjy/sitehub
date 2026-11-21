package icons

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"sitehub/common/protoconv"
	"sitehub/common/rpcx"
	"sitehub/service/rbac/api/internal/svc"
	"sitehub/service/rbac/api/internal/types"
	pb "sitehub/service/rbac/rpc/pb/rbac"
)

func IconContent(ctx context.Context, services *svc.ServiceContext, req *types.IconContentRequest) (string, error) {
	in := &pb.IconName{}
	if err := protoconv.Encode(req, in); err != nil {
		return "", status.Error(codes.InvalidArgument, "invalid request")
	}
	out, err := services.RPC.IconContent(rpcx.Outgoing(ctx, services.Config.ServiceKey), in)
	if err != nil {
		return "", err
	}
	return out.Svg, nil
}
