package logic

import (
	"context"
	"sitehub/service/blog/rpc/internal/svc"
	pb "sitehub/service/blog/rpc/pb/blog"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetArticleHistoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetArticleHistoryLogic(ctx context.Context, s *svc.ServiceContext) *GetArticleHistoryLogic {
	return &GetArticleHistoryLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *GetArticleHistoryLogic) GetArticleHistory(in *pb.IDRequest) (*pb.ArticleHistoryReply, error) {
	actor, err := l.svcCtx.Require(l.ctx, "blog:article:list", "admin", false)
	if err != nil {
		return nil, err
	}
	_ = actor
	items, err := l.svcCtx.Core.Articles.History(articleActor(actor), in.Id)
	return reply[pb.ArticleHistoryReply](map[string]any{"items": items}, err)
}
