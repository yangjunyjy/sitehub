package logic

import (
	"context"
	"sitehub/service/file/rpc/internal/svc"
	pb "sitehub/service/file/rpc/pb/file"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeletePrivateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeletePrivateLogic(ctx context.Context, s *svc.ServiceContext) *DeletePrivateLogic {
	return &DeletePrivateLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *DeletePrivateLogic) DeletePrivate(in *pb.SnapshotRequest) (*pb.Empty, error) {
	return done(l.svcCtx.Core.Private.DeletePrivate(l.ctx, in.Snapshot))
}
