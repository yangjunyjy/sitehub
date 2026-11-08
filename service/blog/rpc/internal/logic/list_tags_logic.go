package logic

import (
	"context"
	"sitehub/service/blog/rpc/internal/svc"
	pb "sitehub/service/blog/rpc/pb/blog"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListTagsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListTagsLogic(ctx context.Context, s *svc.ServiceContext) *ListTagsLogic {
	return &ListTagsLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *ListTagsLogic) ListTags(in *pb.TagListQueryRequest) (*pb.TagPage, error) {
	items, total, err := l.svcCtx.Core.Tags.ListTags(int(in.Page), int(in.PageSize), in.Name)
	return reply[pb.TagPage](map[string]any{"items": items, "total": total}, err)
}
