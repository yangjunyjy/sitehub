package logic

import (
	"context"
	"sitehub/service/product/rpc/internal/svc"
	pb "sitehub/service/product/rpc/pb/product"

	"github.com/zeromicro/go-zero/core/logx"
)

type ResolveDownloadLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewResolveDownloadLogic(ctx context.Context, s *svc.ServiceContext) *ResolveDownloadLogic {
	return &ResolveDownloadLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *ResolveDownloadLogic) ResolveDownload(in *pb.DownloadRequest) (*pb.ProductDownload, error) {
	actor, err := l.svcCtx.Require(l.ctx, "", "member", false)
	if err != nil {
		return nil, err
	}
	_ = actor
	value, err := l.svcCtx.Core.Content.GetPublishedDownload(in.Id)
	return reply[pb.ProductDownload](value, err)
}
