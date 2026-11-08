package logic

import (
	"context"
	"sitehub/common/protoconv"
	"sitehub/common/rpcx"
	"sitehub/service/rbac/api/internal/svc"
	"sitehub/service/rbac/api/internal/types"
	pb "sitehub/service/rbac/rpc/pb/rbac"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GetUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserLogic(ctx context.Context, s *svc.ServiceContext) *GetUserLogic {
	return &GetUserLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *GetUserLogic) GetUser(req *types.GetUserRequest) (*types.Response, error) {
	in := &pb.IDRequest{}
	if err := protoconv.Encode(req, in); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	out, err := l.svcCtx.RPC.GetUser(rpcx.Outgoing(l.ctx, l.svcCtx.Config.ServiceKey), in)
	if err != nil {
		return nil, err
	}
	value, err := protoconv.Value(out)
	if err != nil {
		return nil, err
	}
	return &types.Response{Code: 0, Message: "ok", Data: value}, nil
}
