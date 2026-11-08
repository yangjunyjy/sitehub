package logic

import (
	"context"
	"sitehub/common/rpcx"
	"sitehub/service/blog/rpc/internal/svc"
	pb "sitehub/service/blog/rpc/pb/blog"

	"github.com/zeromicro/go-zero/core/logx"
)

type PublishScheduledLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPublishScheduledLogic(ctx context.Context, s *svc.ServiceContext) *PublishScheduledLogic {
	return &PublishScheduledLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *PublishScheduledLogic) PublishScheduled(in *pb.Empty) (*pb.CountReply, error) {
	count, err := l.svcCtx.Core.Articles.PublishScheduledArticles(l.ctx)
	return &pb.CountReply{Count: count}, rpcx.Error(err)
}
