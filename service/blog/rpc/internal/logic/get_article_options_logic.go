package logic

import (
	"context"
	"sitehub/service/blog/rpc/internal/svc"
	pb "sitehub/service/blog/rpc/pb/blog"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetArticleOptionsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetArticleOptionsLogic(ctx context.Context, s *svc.ServiceContext) *GetArticleOptionsLogic {
	return &GetArticleOptionsLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *GetArticleOptionsLogic) GetArticleOptions(in *pb.Empty) (*pb.ArticleOptions, error) {
	actor, err := l.svcCtx.Require(l.ctx, "", "admin", false)
	if err != nil {
		return nil, err
	}
	_ = actor
	value, err := l.svcCtx.Core.Articles.EditorOptions(articleActor(actor))
	return reply[pb.ArticleOptions](value, err)
}
