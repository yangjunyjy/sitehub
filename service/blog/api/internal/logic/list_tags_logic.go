package logic

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
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
	return ReplyAs[payload.TagPage](map[string]any{"items": items, "total": total}, err)
}
