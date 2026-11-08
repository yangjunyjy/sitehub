package logic

import (
	"context"
	"sitehub/common/rpcx"
	"sitehub/service/rbac/rpc/internal/svc"
	pb "sitehub/service/rbac/rpc/pb/rbac"

	"github.com/zeromicro/go-zero/core/logx"
)

type HealthLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewHealthLogic(ctx context.Context, s *svc.ServiceContext) *HealthLogic {
	return &HealthLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *HealthLogic) Health(in *pb.HealthRequest) (*pb.HealthResponse, error) {
	pool, err := l.svcCtx.DB.DB()
	if err != nil {
		return nil, rpcx.Error(err)
	}
	if err = pool.PingContext(l.ctx); err != nil {
		return nil, rpcx.Error(err)
	}
	if err = l.svcCtx.Cache.Ping(l.ctx).Err(); err != nil {
		return nil, rpcx.Error(err)
	}
	return &pb.HealthResponse{Service: "rbac-rpc", Status: "ok"}, nil
}
