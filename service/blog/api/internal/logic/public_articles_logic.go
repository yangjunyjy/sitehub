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

type PublicArticlesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPublicArticlesLogic(ctx context.Context, s *svc.ServiceContext) *PublicArticlesLogic {
	return &PublicArticlesLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *PublicArticlesLogic) PublicArticles(req *types.PublicArticlesRequest) (*types.Response, error) {

	var input domain.ArticleListQuery
	if err := decodeRequest(req, &input); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	value, err := l.svcCtx.Core.Articles.BrowsePublic(input)
	return ReplyAs[payload.ArticlePage](value, err)
}
