package logic

import (
	"context"
	"sitehub/service/product/rpc/internal/svc"
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
func (l *DownloadsLogic) Downloads(in *pb.Empty) (*pb.WebsiteDownloads, error) {
	value, err := l.svcCtx.Core.Content.GetPublicDownloads()
	return reply[pb.WebsiteDownloads](value, err)
}
