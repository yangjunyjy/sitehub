// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package reader

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"sitehub/common/rpcx"
	"sitehub/service/blog/api/internal/logic"
	"sitehub/service/blog/api/internal/svc"
	"sitehub/service/blog/api/internal/types"
)

type ArticleInteractionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewArticleInteractionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ArticleInteractionLogic {
	return &ArticleInteractionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ArticleInteractionLogic) ArticleInteraction(req *types.ArticleInteractionQuery) (*types.Response, error) {
	var id uint64
	if rpcx.HTTPValue(l.ctx, "authorization") != "" {
		member, err := l.svcCtx.Member(l.ctx)
		if err != nil && status.Code(err) != codes.Unauthenticated {
			return nil, err
		}
		if err == nil {
			id = member.Id
		}
	}
	return logic.Reply(l.svcCtx.Core.Reader.Interaction(l.ctx, id, req.ArticleID))
}
