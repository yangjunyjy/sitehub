package logic

import (
	"context"
	"sitehub/service/product/rpc/internal/svc"
	pb "sitehub/service/product/rpc/pb/product"

	"github.com/zeromicro/go-zero/core/logx"
)

type DocumentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDocumentLogic(ctx context.Context, s *svc.ServiceContext) *DocumentLogic {
	return &DocumentLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *DocumentLogic) Document(in *pb.DocumentQuery) (*pb.WebsiteDocDocument, error) {
	value, err := l.svcCtx.Core.Content.GetDocument(in.Id, in.Version, in.Mode)
	return reply[pb.WebsiteDocDocument](value, err)
}
