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

type CreateFavoriteFolderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateFavoriteFolderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateFavoriteFolderLogic {
	return &CreateFavoriteFolderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateFavoriteFolderLogic) CreateFavoriteFolder(req *types.CreateFavoriteFolderRequest) (*types.Response, error) {
	member, err := l.svcCtx.Member(l.ctx)
	if err != nil {
		return nil, err
	}
	return logic.Reply(l.svcCtx.Core.Reader.CreateFolder(l.ctx, member.Id, req.Name))
}
