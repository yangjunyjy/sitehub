package logic

import (
	"context"
	"sitehub/common/protoconv"
	"sitehub/common/rpcx"
	"sitehub/service/rbac/api/internal/svc"
	"sitehub/service/rbac/api/internal/types"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// Forward preserves the incoming identity and the browser's lossless uint64 representation.
func Forward[In, Out proto.Message](ctx context.Context, services *svc.ServiceContext, request any, in In,
	call func(context.Context, In, ...grpc.CallOption) (Out, error)) (*types.Response, error) {
	if request != nil {
		if err := protoconv.Encode(request, in); err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid request")
		}
	}
	out, err := call(rpcx.Outgoing(ctx, services.Config.ServiceKey), in)
	if err != nil {
		return nil, err
	}
	data, err := protoconv.Value(out)
	if err != nil {
		return nil, err
	}
	return &types.Response{Code: 0, Message: "ok", Data: data}, nil
}
