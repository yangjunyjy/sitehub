package logic

import (
	"context"
	"sitehub/service/product/rpc/internal/svc"
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
func (l *AdminDownloadsLogic) AdminDownloads(in *pb.Empty) (*pb.WebsiteDownloads, error) {
	actor, err := l.svcCtx.Require(l.ctx, "website:downloads:edit", "admin", false)
	if err != nil {
		return nil, err
	}
	_ = actor
	value, err := l.svcCtx.Core.Content.GetDownloads()
	return reply[pb.WebsiteDownloads](value, err)
}
