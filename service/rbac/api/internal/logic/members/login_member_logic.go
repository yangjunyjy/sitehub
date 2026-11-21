package members

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"sitehub/common/protoconv"
	"sitehub/common/rpcx"
	"sitehub/service/rbac/api/internal/session"
	"sitehub/service/rbac/api/internal/svc"
	"sitehub/service/rbac/api/internal/types"
	pb "sitehub/service/rbac/rpc/pb/rbac"
)

func LoginMember(ctx context.Context, services *svc.ServiceContext, req *types.LoginMemberRequest) (*types.Response, error) {
	in := &pb.LoginMemberInput{}
	if err := protoconv.Encode(req, in); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	out, err := services.RPC.LoginMember(rpcx.Outgoing(ctx, services.Config.ServiceKey), in)
	if err != nil {
		return nil, err
	}
	value, err := protoconv.Value(out)
	if err != nil {
		return nil, err
	}
	session.Issue(ctx, "site", out.GetToken(), out.GetExpiresAt())
	return &types.Response{Code: 0, Message: "ok", Data: value}, nil
}
