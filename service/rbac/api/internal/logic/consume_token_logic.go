// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package logic

import (
	"context"
	"sitehub/common/protoconv"
	"sitehub/common/rpcx"
	pb "sitehub/service/rbac/rpc/pb/rbac"

	"sitehub/service/rbac/api/internal/svc"
	"sitehub/service/rbac/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ConsumeTokenLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewConsumeTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ConsumeTokenLogic {
	return &ConsumeTokenLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ConsumeTokenLogic) ConsumeToken(req *types.ConsumeTokenRequest) (resp *types.Response, err error) {
	out, err := l.svcCtx.RPC.ConsumeToken(rpcx.Outgoing(l.ctx, l.svcCtx.Config.ServiceKey), &pb.ConsumeTokenRequest{Token: req.Token})
	if err != nil {
		return nil, err
	}
	value, err := protoconv.Value(out)
	if err != nil {
		return nil, err
	}
	return &types.Response{Code: 0, Message: "ok", Data: value}, nil
}
