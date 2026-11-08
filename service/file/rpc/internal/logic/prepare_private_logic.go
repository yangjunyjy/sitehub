package logic

import (
	"context"
	"sitehub/common/rpcx"
	"sitehub/service/file/rpc/internal/svc"
	pb "sitehub/service/file/rpc/pb/file"

	"github.com/zeromicro/go-zero/core/logx"
)

type PreparePrivateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPreparePrivateLogic(ctx context.Context, s *svc.ServiceContext) *PreparePrivateLogic {
	return &PreparePrivateLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *PreparePrivateLogic) PreparePrivate(in *pb.SnapshotRequest) (*pb.SnapshotReply, error) {
	value, err := l.svcCtx.Core.Private.PreparePrivate(in.Key)
	return &pb.SnapshotReply{Snapshot: value}, rpcx.Error(err)
}
