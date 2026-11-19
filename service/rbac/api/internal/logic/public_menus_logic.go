package logic

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

type PublicMenusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPublicMenusLogic(ctx context.Context, s *svc.ServiceContext) *PublicMenusLogic {
	return &PublicMenusLogic{ctx, s}
}
func (l *PublicMenusLogic) PublicMenus(req *types.SystemRequest) (*types.Response, error) {
	in := &pb.SystemRequest{}
	if err := protoconv.Encode(req, in); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	out, err := l.svcCtx.RPC.PublicMenus(rpcx.Outgoing(l.ctx, l.svcCtx.Config.ServiceKey), in)
	if err != nil {
		return nil, err
	}
	value, err := protoconv.Value(out)
	if err != nil {
		return nil, err
	}
	return &types.Response{Code: 0, Message: "ok", Data: value}, nil
}
