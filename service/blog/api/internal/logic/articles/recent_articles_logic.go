package articles

import (
	"context"
	logicshared "sitehub/service/blog/api/internal/logic"
	"sitehub/service/blog/api/internal/payload"
	"sitehub/service/blog/api/internal/svc"
	"sitehub/service/blog/api/internal/types"
	"sitehub/service/blog/domain"

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
func (l *RecentArticlesLogic) RecentArticles(req *types.RecentArticlesRequest) (*types.Response, error) {

	value, err := l.svcCtx.Core.Articles.BrowsePublic(domain.ArticleListQuery{Page: 1, PageSize: int(req.Limit), Sort: "recent"})
	return logicshared.ReplyAs[payload.ArticlePage](value, err)
}
