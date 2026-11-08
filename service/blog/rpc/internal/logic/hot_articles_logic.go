package logic

import (
	"context"
	domain "sitehub/service/blog/domain"
	"sitehub/service/blog/rpc/internal/svc"
	pb "sitehub/service/blog/rpc/pb/blog"

	"github.com/zeromicro/go-zero/core/logx"
)

type HotArticlesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewHotArticlesLogic(ctx context.Context, s *svc.ServiceContext) *HotArticlesLogic {
	return &HotArticlesLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *HotArticlesLogic) HotArticles(in *pb.LimitRequest) (*pb.ArticlePage, error) {
	value, err := l.svcCtx.Core.Articles.BrowsePublic(domain.ArticleListQuery{Page: 1, PageSize: int(in.Limit), Sort: "hot"})
	return reply[pb.ArticlePage](value, err)
}
