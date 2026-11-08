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

type CurrentMenusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCurrentMenusLogic(ctx context.Context, s *svc.ServiceContext) *CurrentMenusLogic {
	return &CurrentMenusLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *CurrentMenusLogic) CurrentMenus() (*types.Response, error) {
	in := &pb.Empty{}
	out, err := l.svcCtx.RPC.CurrentMenus(rpcx.Outgoing(l.ctx, l.svcCtx.Config.ServiceKey), in)
	if err != nil {
		return nil, err
	}
	value, err := protoconv.Value(out)
	if err != nil {
		return nil, err
	}
	return &types.Response{Code: 0, Message: "ok", Data: value}, nil
}
