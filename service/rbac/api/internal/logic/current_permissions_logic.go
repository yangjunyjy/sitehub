package logic

import (
	"context"
	"sitehub/common/protoconv"
	"sitehub/common/rpcx"
	"sitehub/service/rbac/api/internal/svc"
	"sitehub/service/rbac/api/internal/types"
	pb "sitehub/service/rbac/rpc/pb/rbac"

	"github.com/zeromicro/go-zero/core/logx"
)

type CurrentPermissionsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCurrentPermissionsLogic(ctx context.Context, s *svc.ServiceContext) *CurrentPermissionsLogic {
	return &CurrentPermissionsLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *CurrentPermissionsLogic) CurrentPermissions() (*types.Response, error) {
	in := &pb.Empty{}
	out, err := l.svcCtx.RPC.CurrentPermissions(rpcx.Outgoing(l.ctx, l.svcCtx.Config.ServiceKey), in)
	if err != nil {
		return nil, err
	}
	value, err := protoconv.Value(out)
	if err != nil {
		return nil, err
	}
	return &types.Response{Code: 0, Message: "ok", Data: value}, nil
}
