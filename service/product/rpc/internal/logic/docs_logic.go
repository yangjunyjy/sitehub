package logic

import (
	"context"
	"sitehub/service/product/rpc/internal/svc"
	pb "sitehub/service/product/rpc/pb/product"

	"github.com/zeromicro/go-zero/core/logx"
)

type DocsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDocsLogic(ctx context.Context, s *svc.ServiceContext) *DocsLogic {
	return &DocsLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *DocsLogic) Docs(in *pb.Empty) (*pb.WebsiteDocs, error) {
	actor, err := l.svcCtx.Require(l.ctx, "website:docs:edit", "admin", false)
	if err != nil {
		return nil, err
	}
	_ = actor
	value, err := l.svcCtx.Core.Content.GetDocs()
	return reply[pb.WebsiteDocs](value, err)
}
