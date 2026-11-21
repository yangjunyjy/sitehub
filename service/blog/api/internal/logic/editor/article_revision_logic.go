package editor

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	logicshared "sitehub/service/blog/api/internal/logic"
	"sitehub/service/blog/api/internal/payload"
	"sitehub/service/blog/api/internal/svc"
	"sitehub/service/blog/api/internal/types"
)

type ArticleRevisionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewArticleRevisionLogic(ctx context.Context, s *svc.ServiceContext) *ArticleRevisionLogic {
	return &ArticleRevisionLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *ArticleRevisionLogic) ArticleRevision(req *types.ArticleRevisionRequest) (*types.Response, error) {

	actor, err := l.svcCtx.Require(l.ctx, "blog:article:list", "admin", false)
	if err != nil {
		return nil, err
	}
	value, err := l.svcCtx.Core.Articles.RevisionDetail(logicshared.ArticleActor(actor), req.ID, req.RevisionID)
	return logicshared.ReplyAs[payload.ArticleView](value, err)
}
