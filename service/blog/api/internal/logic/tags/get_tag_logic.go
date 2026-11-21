package tags

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	logicshared "sitehub/service/blog/api/internal/logic"
	"sitehub/service/blog/api/internal/payload"
	"sitehub/service/blog/api/internal/svc"
	"sitehub/service/blog/api/internal/types"
)

type GetTagLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetTagLogic(ctx context.Context, s *svc.ServiceContext) *GetTagLogic {
	return &GetTagLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *GetTagLogic) GetTag(req *types.GetTagRequest) (*types.Response, error) {

	value, err := l.svcCtx.Core.Tags.GetTag(uint(req.ID))
	return logicshared.ReplyAs[payload.Tag](value, err)
}
