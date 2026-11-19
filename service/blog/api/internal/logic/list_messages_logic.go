package logic

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sitehub/service/blog/api/internal/svc"
	"sitehub/service/blog/api/internal/types"
	"strconv"
)

type ListMessagesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListMessagesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMessagesLogic {
	return &ListMessagesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListMessagesLogic) ListMessages(req *types.CommunityPage) (*types.Response, error) {

	page, size := req.Page, req.PageSize
	if page == 0 {
		page = 1
	}
	if size == 0 {
		size = 40
	}
	if page < 1 || page > 1000000 || size < 1 || size > 100 {
		return nil, status.Error(codes.InvalidArgument, "无效的分页参数")
	}
	messages, err := l.svcCtx.Core.Community.PublishedMessages(l.ctx, int((page-1)*size), int(size))
	if err != nil {
		return nil, err
	}
	items := make([]map[string]any, 0, len(messages))
	for _, message := range messages {
		items = append(items, map[string]any{"id": strconv.FormatUint(message.ID, 10), "nickname": message.Nickname, "content": message.Content})
	}
	return reply(map[string]any{"items": items}, nil)
}
