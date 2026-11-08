package logic

import (
	"context"
	"sitehub/common/rpcx"
	"sitehub/service/file/rpc/internal/svc"
	pb "sitehub/service/file/rpc/pb/file"

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
	return &pb.HealthResponse{Service: "file-rpc", Status: "ok"}, nil
}
