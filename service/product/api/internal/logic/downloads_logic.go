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

type DownloadsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDownloadsLogic(ctx context.Context, s *svc.ServiceContext) *DownloadsLogic {
	return &DownloadsLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *DownloadsLogic) Downloads() (*types.Response, error) {
	in := &pb.Empty{}
	out, err := l.svcCtx.RPC.Downloads(rpcx.Outgoing(l.ctx, l.svcCtx.Config.ServiceKey), in)
	if err != nil {
		return nil, err
	}
	value, err := protoconv.Value(out)
	if err != nil {
		return nil, err
	}
	return &types.Response{Code: 0, Message: "ok", Data: value}, nil
}
