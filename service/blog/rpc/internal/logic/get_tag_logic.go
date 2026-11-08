package logic

import (
	"context"
	"sitehub/service/blog/rpc/internal/svc"
	pb "sitehub/service/blog/rpc/pb/blog"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTagLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetTagLogic(ctx context.Context, s *svc.ServiceContext) *GetTagLogic {
	return &GetTagLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *GetTagLogic) GetTag(in *pb.IDRequest) (*pb.EntityTag, error) {
	value, err := l.svcCtx.Core.Tags.GetTag(uint(in.Id))
	return reply[pb.EntityTag](value, err)
}
