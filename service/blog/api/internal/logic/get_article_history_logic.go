package logic

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	"sitehub/service/blog/api/internal/payload"
	"sitehub/service/blog/api/internal/svc"
	"sitehub/service/blog/api/internal/types"
)

type GetArticleHistoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetArticleHistoryLogic(ctx context.Context, s *svc.ServiceContext) *GetArticleHistoryLogic {
	return &GetArticleHistoryLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *GetArticleHistoryLogic) GetArticleHistory(req *types.GetArticleHistoryRequest) (*types.Response, error) {

	actor, err := l.svcCtx.Require(l.ctx, "blog:article:list", "admin", false)
	if err != nil {
		return nil, err
	}
	items, err := l.svcCtx.Core.Articles.History(ArticleActor(actor), req.ID)
	return ReplyAs[payload.ArticleHistoryList](map[string]any{"items": items}, err)
}
