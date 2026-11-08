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

type AdminDownloadsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminDownloadsLogic(ctx context.Context, s *svc.ServiceContext) *AdminDownloadsLogic {
	return &AdminDownloadsLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *AdminDownloadsLogic) AdminDownloads() (*types.Response, error) {
	in := &pb.Empty{}
	out, err := l.svcCtx.RPC.AdminDownloads(rpcx.Outgoing(l.ctx, l.svcCtx.Config.ServiceKey), in)
	if err != nil {
		return nil, err
	}
	value, err := protoconv.Value(out)
	if err != nil {
		return nil, err
	}
	return &types.Response{Code: 0, Message: "ok", Data: value}, nil
}
