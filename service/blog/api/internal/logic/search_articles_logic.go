package logic

import (
	"context"
	"sitehub/service/blog/api/internal/payload"
	"sitehub/service/blog/api/internal/svc"
	"sitehub/service/blog/api/internal/types"
	"sitehub/service/blog/domain"

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
func (l *SearchArticlesLogic) SearchArticles(req *types.SearchArticlesRequest) (*types.Response, error) {

	value, err := l.svcCtx.Core.Articles.BrowsePublic(domain.ArticleListQuery{Page: int(req.Page), PageSize: int(req.PageSize), Title: req.Keyword})
	return ReplyAs[payload.ArticlePage](value, err)
}
