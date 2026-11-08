package logic

import (
	"context"
	"sitehub/service/blog/rpc/internal/svc"
	pb "sitehub/service/blog/rpc/pb/blog"

	"github.com/zeromicro/go-zero/core/logx"
)

type ArticleRevisionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewArticleRevisionLogic(ctx context.Context, s *svc.ServiceContext) *ArticleRevisionLogic {
	return &ArticleRevisionLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *ArticleRevisionLogic) ArticleRevision(in *pb.ArticleRevisionRequest) (*pb.ArticleView, error) {
	actor, err := l.svcCtx.Require(l.ctx, "blog:article:list", "admin", false)
	if err != nil {
		return nil, err
	}
	_ = actor
	value, err := l.svcCtx.Core.Articles.RevisionDetail(articleActor(actor), in.Id, in.RevisionId)
	return reply[pb.ArticleView](value, err)
}
