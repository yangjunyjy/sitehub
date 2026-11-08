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

type AdminArticlesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminArticlesLogic(ctx context.Context, s *svc.ServiceContext) *AdminArticlesLogic {
	return &AdminArticlesLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *AdminArticlesLogic) AdminArticles(in *pb.ArticleListQuery) (*pb.ArticlePage, error) {
	actor, err := l.svcCtx.Require(l.ctx, "blog:article:list", "admin", false)
	if err != nil {
		return nil, err
	}
	_ = actor
	var input domain.ArticleListQuery
	if err := protoconv.Decode(in, &input); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	value, err := l.svcCtx.Core.Articles.BrowseEditor(articleActor(actor), input)
	return reply[pb.ArticlePage](value, err)
}
