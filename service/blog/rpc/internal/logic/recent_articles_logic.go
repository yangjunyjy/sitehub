package logic

import (
	"context"
	domain "sitehub/service/blog/domain"
	"sitehub/service/blog/rpc/internal/svc"
	pb "sitehub/service/blog/rpc/pb/blog"

	"github.com/zeromicro/go-zero/core/logx"
)

type RecentArticlesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRecentArticlesLogic(ctx context.Context, s *svc.ServiceContext) *RecentArticlesLogic {
	return &RecentArticlesLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *RecentArticlesLogic) RecentArticles(in *pb.LimitRequest) (*pb.ArticlePage, error) {
	value, err := l.svcCtx.Core.Articles.BrowsePublic(domain.ArticleListQuery{Page: 1, PageSize: int(in.Limit), Sort: "recent"})
	return reply[pb.ArticlePage](value, err)
}
