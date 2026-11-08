package logic

import (
	"context"
	"sitehub/common/protoconv"
	"sitehub/common/rpcx"
	"sitehub/service/blog/api/internal/svc"
	"sitehub/service/blog/api/internal/types"
	pb "sitehub/service/blog/rpc/pb/blog"

	"github.com/zeromicro/go-zero/core/logx"
)

type DashboardLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDashboardLogic(ctx context.Context, s *svc.ServiceContext) *DashboardLogic {
	return &DashboardLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *DashboardLogic) Dashboard() (*types.Response, error) {
	in := &pb.Empty{}
	out, err := l.svcCtx.RPC.Dashboard(rpcx.Outgoing(l.ctx, l.svcCtx.Config.ServiceKey), in)
	if err != nil {
		return nil, err
	}
	value, err := protoconv.Value(out)
	if err != nil {
		return nil, err
	}
	return &types.Response{Code: 0, Message: "ok", Data: value}, nil
}
