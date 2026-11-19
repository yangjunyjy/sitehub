package logic

import (
	"context"
	"sitehub/common/protoconv"
	"sitehub/common/rpcx"
	"sitehub/service/rbac/api/internal/svc"
	"sitehub/service/rbac/api/internal/types"
	pb "sitehub/service/rbac/rpc/pb/rbac"
)

type ListSystemsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListSystemsLogic(ctx context.Context, s *svc.ServiceContext) *ListSystemsLogic {
	return &ListSystemsLogic{ctx, s}
}
func (l *ListSystemsLogic) ListSystems() (*types.Response, error) {
	in := &pb.Empty{}
	out, err := l.svcCtx.RPC.ListSystems(rpcx.Outgoing(l.ctx, l.svcCtx.Config.ServiceKey), in)
	if err != nil {
		return nil, err
	}
	value, err := protoconv.Value(out)
	if err != nil {
		return nil, err
	}
	return &types.Response{Code: 0, Message: "ok", Data: value}, nil
}
