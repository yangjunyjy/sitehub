package logic

import (
	"context"
	"sitehub/service/blog/rpc/internal/svc"
	pb "sitehub/service/blog/rpc/pb/blog"

	"github.com/zeromicro/go-zero/core/logx"
)

type CleanupAssetsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCleanupAssetsLogic(ctx context.Context, s *svc.ServiceContext) *CleanupAssetsLogic {
	return &CleanupAssetsLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *CleanupAssetsLogic) CleanupAssets(in *pb.Empty) (*pb.Empty, error) {
	return done(l.svcCtx.Core.Assets.CleanupArticleAssets(l.ctx))
}
