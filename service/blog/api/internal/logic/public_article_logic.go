package logic

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	"sitehub/service/blog/api/internal/payload"
	"sitehub/service/blog/api/internal/svc"
	"sitehub/service/blog/api/internal/types"
)

type PublicArticleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPublicArticleLogic(ctx context.Context, s *svc.ServiceContext) *PublicArticleLogic {
	return &PublicArticleLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *PublicArticleLogic) PublicArticle(req *types.PublicArticleRequest) (*types.Response, error) {

	value, err := l.svcCtx.Core.Articles.PublicDetail(req.ID)
	return ReplyAs[payload.ArticleView](value, err)
}
