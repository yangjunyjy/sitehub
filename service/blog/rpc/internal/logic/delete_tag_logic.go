package logic

import (
	"context"
	"sitehub/service/blog/rpc/internal/svc"
	pb "sitehub/service/blog/rpc/pb/blog"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteTagLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteTagLogic(ctx context.Context, s *svc.ServiceContext) *DeleteTagLogic {
	return &DeleteTagLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *DeleteTagLogic) DeleteTag(in *pb.IDRequest) (*pb.Empty, error) {
	actor, err := l.svcCtx.Require(l.ctx, "blog:tag:delete", "admin", true)
	if err != nil {
		return nil, err
	}
	_ = actor

	err = l.svcCtx.Core.Tags.DeleteTag(uint(in.Id))
	return done(err)
}
