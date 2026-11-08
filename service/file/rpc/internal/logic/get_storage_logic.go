package logic

import (
	"context"
	"sitehub/service/file/rpc/internal/svc"
	pb "sitehub/service/file/rpc/pb/file"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetStorageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetStorageLogic(ctx context.Context, s *svc.ServiceContext) *GetStorageLogic {
	return &GetStorageLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *GetStorageLogic) GetStorage(in *pb.Empty) (*pb.WebsiteStorage, error) {
	actor, err := l.svcCtx.Require(l.ctx, "website:storage:edit", "admin", false)
	if err != nil {
		return nil, err
	}
	_ = actor
	value, err := l.svcCtx.Core.Settings.GetStorage()
	value.SecretKey = ""
	return reply[pb.WebsiteStorage](value, err)
}
