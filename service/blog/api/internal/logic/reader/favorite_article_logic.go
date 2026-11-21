// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package reader

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"sitehub/service/blog/api/internal/logic"
	"sitehub/service/blog/api/internal/svc"
	"sitehub/service/blog/api/internal/types"
)

type FavoriteArticleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFavoriteArticleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FavoriteArticleLogic {
	return &FavoriteArticleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FavoriteArticleLogic) FavoriteArticle(req *types.FavoriteArticleRequest) (*types.Response, error) {
	member, err := l.svcCtx.Member(l.ctx)
	if err != nil {
		return nil, err
	}
	return logic.Reply(l.svcCtx.Core.Reader.SetFavorite(l.ctx, member.Id, req.ArticleID, req.FolderID, req.Favorited))
}
