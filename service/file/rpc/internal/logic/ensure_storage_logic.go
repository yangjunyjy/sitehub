package logic

import (
	"context"
	"sitehub/service/file/rpc/internal/svc"
	pb "sitehub/service/file/rpc/pb/file"

	"github.com/zeromicro/go-zero/core/logx"
)

type EnsureStorageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewEnsureStorageLogic(ctx context.Context, s *svc.ServiceContext) *EnsureStorageLogic {
	return &EnsureStorageLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *EnsureStorageLogic) EnsureStorage(in *pb.Empty) (*pb.Empty, error) {
	return done(l.svcCtx.Core.Settings.EnsureDefaults())
}
