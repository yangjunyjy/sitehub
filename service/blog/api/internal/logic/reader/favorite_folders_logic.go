// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package reader

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"sitehub/service/blog/api/internal/logic"
	"sitehub/service/blog/api/internal/svc"
	"sitehub/service/blog/api/internal/types"
)

type FavoriteFoldersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFavoriteFoldersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FavoriteFoldersLogic {
	return &FavoriteFoldersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FavoriteFoldersLogic) FavoriteFolders() (*types.Response, error) {
	member, err := l.svcCtx.Member(l.ctx)
	if err != nil {
		return nil, err
	}
	items, err := l.svcCtx.Core.Reader.Folders(l.ctx, member.Id)
	return logic.Reply(map[string]any{"items": items}, err)
}
