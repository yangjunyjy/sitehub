package logic

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	"sitehub/service/blog/api/internal/payload"
	"sitehub/service/blog/api/internal/svc"
	"sitehub/service/blog/api/internal/types"
)

type GetArticleOptionsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetArticleOptionsLogic(ctx context.Context, s *svc.ServiceContext) *GetArticleOptionsLogic {
	return &GetArticleOptionsLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *GetArticleOptionsLogic) GetArticleOptions() (*types.Response, error) {

	actor, err := l.svcCtx.Require(l.ctx, "", "admin", false)
	if err != nil {
		return nil, err
	}
	value, err := l.svcCtx.Core.Articles.EditorOptions(ArticleActor(actor))
	return ReplyAs[payload.ArticleOptions](value, err)
}
