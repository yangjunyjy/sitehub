package logic

import (
	"context"
	"sitehub/service/product/rpc/internal/svc"
	pb "sitehub/service/product/rpc/pb/product"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminDocumentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminDocumentLogic(ctx context.Context, s *svc.ServiceContext) *AdminDocumentLogic {
	return &AdminDocumentLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *AdminDocumentLogic) AdminDocument(in *pb.DocumentQuery) (*pb.WebsiteDocDocument, error) {
	actor, err := l.svcCtx.Require(l.ctx, "website:docs:edit", "admin", false)
	if err != nil {
		return nil, err
	}
	_ = actor
	value, err := l.svcCtx.Core.Content.GetAdminDocument(in.Id, in.Version)
	return reply[pb.WebsiteDocDocument](value, err)
}
