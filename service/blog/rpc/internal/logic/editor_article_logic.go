package logic

import (
	"context"
	"sitehub/service/blog/rpc/internal/svc"
	pb "sitehub/service/blog/rpc/pb/blog"

	"github.com/zeromicro/go-zero/core/logx"
)

type EditorArticleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewEditorArticleLogic(ctx context.Context, s *svc.ServiceContext) *EditorArticleLogic {
	return &EditorArticleLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *EditorArticleLogic) EditorArticle(in *pb.IDRequest) (*pb.ArticleView, error) {
	actor, err := l.svcCtx.Require(l.ctx, "blog:article:list", "admin", false)
	if err != nil {
		return nil, err
	}
	_ = actor
	value, err := l.svcCtx.Core.Articles.EditorDetail(articleActor(actor), in.Id)
	return reply[pb.ArticleView](value, err)
}
