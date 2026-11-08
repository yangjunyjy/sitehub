package logic

import (
	"context"
	"sitehub/service/file/rpc/internal/svc"
	pb "sitehub/service/file/rpc/pb/file"

	"github.com/zeromicro/go-zero/core/logx"
)

type StorageSnapshotLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewStorageSnapshotLogic(ctx context.Context, s *svc.ServiceContext) *StorageSnapshotLogic {
	return &StorageSnapshotLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *StorageSnapshotLogic) StorageSnapshot(in *pb.Empty) (*pb.WebsiteStorage, error) {
	value, err := l.svcCtx.Core.Settings.GetStorage()
	return reply[pb.WebsiteStorage](value, err)
}
