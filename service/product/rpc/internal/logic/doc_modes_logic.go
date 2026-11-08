package logic

import (
	"context"
	"sitehub/service/product/rpc/internal/svc"
	pb "sitehub/service/product/rpc/pb/product"

	"github.com/zeromicro/go-zero/core/logx"
)

type DocModesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDocModesLogic(ctx context.Context, s *svc.ServiceContext) *DocModesLogic {
	return &DocModesLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *DocModesLogic) DocModes(in *pb.Empty) (*pb.WebsiteDocModes, error) {
	value, err := l.svcCtx.Core.Content.GetDocModes()
	return reply[pb.WebsiteDocModes](value, err)
}
