package logic

import (
	"context"
	domain "sitehub/service/blog/domain"
	"sitehub/service/blog/rpc/internal/svc"
	pb "sitehub/service/blog/rpc/pb/blog"

	"github.com/zeromicro/go-zero/core/logx"
)

type SearchArticlesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSearchArticlesLogic(ctx context.Context, s *svc.ServiceContext) *SearchArticlesLogic {
	return &SearchArticlesLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *SearchArticlesLogic) SearchArticles(in *pb.SearchRequest) (*pb.ArticlePage, error) {
	value, err := l.svcCtx.Core.Articles.BrowsePublic(domain.ArticleListQuery{Page: int(in.Page), PageSize: int(in.PageSize), Title: in.Keyword})
	return reply[pb.ArticlePage](value, err)
}
