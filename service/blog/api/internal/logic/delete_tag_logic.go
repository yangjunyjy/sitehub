package logic

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	"sitehub/service/blog/api/internal/svc"
	"sitehub/service/blog/api/internal/types"
)

type DeleteTagLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteTagLogic(ctx context.Context, s *svc.ServiceContext) *DeleteTagLogic {
	return &DeleteTagLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *DeleteTagLogic) DeleteTag(req *types.DeleteTagRequest) (*types.Response, error) {

	_, err := l.svcCtx.Require(l.ctx, "blog:tag:delete", "admin", true)
	if err != nil {
		return nil, err
	}

	err = l.svcCtx.Core.Tags.DeleteTag(uint(req.ID))
	return done(err)
}
