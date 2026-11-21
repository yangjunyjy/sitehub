package editor

import (
	"context"
	logicshared "sitehub/service/blog/api/internal/logic"
	"sitehub/service/blog/api/internal/payload"
	"sitehub/service/blog/api/internal/svc"
	"sitehub/service/blog/api/internal/types"
	"sitehub/service/blog/domain"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ApplyArticleActionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewApplyArticleActionLogic(ctx context.Context, s *svc.ServiceContext) *ApplyArticleActionLogic {
	return &ApplyArticleActionLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *ApplyArticleActionLogic) ApplyArticleAction(req *types.ApplyArticleActionRequest) (*types.Response, error) {

	actor, err := l.svcCtx.Require(l.ctx, "", "admin", true)
	if err != nil {
		return nil, err
	}
	var input domain.ArticleAction
	if err := logicshared.DecodeRequest(req, &input); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	value, err := l.svcCtx.Core.Articles.ActOnArticle(logicshared.ArticleActor(actor), input)
	return logicshared.ReplyAs[payload.ArticleView](value, err)
}
