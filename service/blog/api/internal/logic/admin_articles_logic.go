package logic

import (
	"context"
	"sitehub/service/blog/api/internal/payload"
	"sitehub/service/blog/api/internal/svc"
	"sitehub/service/blog/api/internal/types"
	"sitehub/service/blog/domain"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AdminArticlesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminArticlesLogic(ctx context.Context, s *svc.ServiceContext) *AdminArticlesLogic {
	return &AdminArticlesLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *AdminArticlesLogic) AdminArticles(req *types.AdminArticlesRequest) (*types.Response, error) {

	actor, err := l.svcCtx.Require(l.ctx, "blog:article:list", "admin", false)
	if err != nil {
		return nil, err
	}
	var input domain.ArticleListQuery
	if err := decodeRequest(req, &input); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	value, err := l.svcCtx.Core.Articles.BrowseEditor(ArticleActor(actor), input)
	return ReplyAs[payload.ArticlePage](value, err)
}
