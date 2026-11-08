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

type SaveDraftLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSaveDraftLogic(ctx context.Context, s *svc.ServiceContext) *SaveDraftLogic {
	return &SaveDraftLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *SaveDraftLogic) SaveDraft(in *pb.SaveArticle) (*pb.ArticleView, error) {
	actor, err := l.svcCtx.Require(l.ctx, "", "admin", true)
	if err != nil {
		return nil, err
	}
	_ = actor
	var input domain.SaveArticle
	if err := protoconv.Decode(in, &input); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	value, err := l.svcCtx.Core.Articles.SaveDraft(articleActor(actor), input)
	return reply[pb.ArticleView](value, err)
}
