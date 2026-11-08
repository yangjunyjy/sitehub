package logic

import (
	"context"
	"sitehub/service/blog/rpc/internal/svc"
	pb "sitehub/service/blog/rpc/pb/blog"

	"github.com/zeromicro/go-zero/core/logx"
)

type HotTagsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewHotTagsLogic(ctx context.Context, s *svc.ServiceContext) *HotTagsLogic {
	return &HotTagsLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *HotTagsLogic) HotTags(in *pb.LimitRequest) (*pb.Records, error) {
	actor, err := l.svcCtx.Require(l.ctx, "blog:article:list", "admin", false)
	if err != nil {
		return nil, err
	}
	_ = actor
	items, err := l.svcCtx.Core.Statistics.GetHotTags(int(in.Limit))
	return reply[pb.Records](map[string]any{"items": items}, err)
}
