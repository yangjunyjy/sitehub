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

type HotArticlesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewHotArticlesLogic(ctx context.Context, s *svc.ServiceContext) *HotArticlesLogic {
	return &HotArticlesLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *HotArticlesLogic) HotArticles(req *types.HotArticlesRequest) (*types.Response, error) {

	value, err := l.svcCtx.Core.Articles.BrowsePublic(domain.ArticleListQuery{Page: 1, PageSize: int(req.Limit), Sort: "hot"})
	return logicshared.ReplyAs[payload.ArticlePage](value, err)
}
