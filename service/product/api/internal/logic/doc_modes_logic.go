package logic

import (
	"context"
	"sitehub/common/protoconv"
	"sitehub/common/rpcx"
	"sitehub/service/product/api/internal/svc"
	"sitehub/service/product/api/internal/types"
	pb "sitehub/service/product/rpc/pb/product"

	"github.com/zeromicro/go-zero/core/logx"
)

type DocModesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDocModesLogic(ctx context.Context, s *svc.ServiceContext) *DocModesLogic {
	return &DocModesLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *DocModesLogic) DocModes() (*types.Response, error) {
	in := &pb.Empty{}
	out, err := l.svcCtx.RPC.DocModes(rpcx.Outgoing(l.ctx, l.svcCtx.Config.ServiceKey), in)
	if err != nil {
		return nil, err
	}
	value, err := protoconv.Value(out)
	if err != nil {
		return nil, err
	}
	return &types.Response{Code: 0, Message: "ok", Data: value}, nil
}
