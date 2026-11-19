package logic

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	"sitehub/service/blog/api/internal/svc"
	"sitehub/service/blog/api/internal/types"
	"strconv"
)

type ListFriendsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListFriendsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListFriendsLogic {
	return &ListFriendsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListFriendsLogic) ListFriends() (*types.Response, error) {

	links, err := l.svcCtx.Core.Community.PublishedFriends(l.ctx)
	if err != nil {
		return nil, err
	}
	items := make([]map[string]any, 0, len(links))
	for _, link := range links {
		items = append(items, map[string]any{"id": strconv.FormatUint(link.ID, 10), "name": link.Name, "url": link.URL, "description": link.Description, "avatar": link.Avatar})
	}
	return reply(map[string]any{"items": items}, nil)
}
