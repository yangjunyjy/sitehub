package tags

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	logicshared "sitehub/service/blog/api/internal/logic"
	"sitehub/service/blog/api/internal/payload"
	"sitehub/service/blog/api/internal/svc"
	"sitehub/service/blog/api/internal/types"
)

type ListTagsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListTagsLogic(ctx context.Context, s *svc.ServiceContext) *ListTagsLogic {
	return &ListTagsLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *ListTagsLogic) ListTags(req *types.ListTagsRequest) (*types.Response, error) {

	items, total, err := l.svcCtx.Core.Tags.ListTags(int(req.Page), int(req.PageSize), req.Name)
	return logicshared.ReplyAs[payload.TagPage](map[string]any{"items": items, "total": total}, err)
}
