package editor

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	logicshared "sitehub/service/blog/api/internal/logic"
	"sitehub/service/blog/api/internal/payload"
	"sitehub/service/blog/api/internal/svc"
	"sitehub/service/blog/api/internal/types"
)

type EditorArticleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewEditorArticleLogic(ctx context.Context, s *svc.ServiceContext) *EditorArticleLogic {
	return &EditorArticleLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *EditorArticleLogic) EditorArticle(req *types.EditorArticleRequest) (*types.Response, error) {

	actor, err := l.svcCtx.Require(l.ctx, "blog:article:list", "admin", false)
	if err != nil {
		return nil, err
	}
	value, err := l.svcCtx.Core.Articles.EditorDetail(logicshared.ArticleActor(actor), req.ID)
	return logicshared.ReplyAs[payload.ArticleView](value, err)
}
