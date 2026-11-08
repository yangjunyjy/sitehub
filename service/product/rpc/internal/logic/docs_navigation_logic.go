package logic

import (
	"context"
	"sitehub/service/product/rpc/internal/svc"
	pb "sitehub/service/product/rpc/pb/product"

	"github.com/zeromicro/go-zero/core/logx"
)

type DocsNavigationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDocsNavigationLogic(ctx context.Context, s *svc.ServiceContext) *DocsNavigationLogic {
	return &DocsNavigationLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *DocsNavigationLogic) DocsNavigation(in *pb.DocumentQuery) (*pb.WebsiteDocsNavigation, error) {
	value, err := l.svcCtx.Core.Content.GetDocsNavigation(in.Version, in.Mode)
	return reply[pb.WebsiteDocsNavigation](value, err)
}
