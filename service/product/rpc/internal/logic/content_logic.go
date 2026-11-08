package logic

import (
	"context"
	"sitehub/service/product/rpc/internal/svc"
	pb "sitehub/service/product/rpc/pb/product"

	"github.com/zeromicro/go-zero/core/logx"
)

type ContentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewContentLogic(ctx context.Context, s *svc.ServiceContext) *ContentLogic {
	return &ContentLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *ContentLogic) Content(in *pb.Empty) (*pb.WebsiteContent, error) {
	value, err := l.svcCtx.Core.Content.GetContent()
	return reply[pb.WebsiteContent](value, err)
}
