package logic

import (
	"context"
	"sitehub/service/blog/rpc/internal/svc"
	pb "sitehub/service/blog/rpc/pb/blog"

	"github.com/zeromicro/go-zero/core/logx"
)

type PublicArticleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPublicArticleLogic(ctx context.Context, s *svc.ServiceContext) *PublicArticleLogic {
	return &PublicArticleLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *PublicArticleLogic) PublicArticle(in *pb.IDRequest) (*pb.ArticleView, error) {
	value, err := l.svcCtx.Core.Articles.PublicDetail(in.Id)
	return reply[pb.ArticleView](value, err)
}
