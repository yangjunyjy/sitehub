package logic

import (
	"context"
	"sitehub/common/protoconv"
	domain "sitehub/service/blog/domain"
	"sitehub/service/blog/rpc/internal/svc"
	pb "sitehub/service/blog/rpc/pb/blog"

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
func (l *PublicArticlesLogic) PublicArticles(in *pb.ArticleListQuery) (*pb.ArticlePage, error) {
	var input domain.ArticleListQuery
	if err := protoconv.Decode(in, &input); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	value, err := l.svcCtx.Core.Articles.BrowsePublic(input)
	return reply[pb.ArticlePage](value, err)
}
