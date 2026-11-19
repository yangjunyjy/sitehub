package logic

import (
	"context"
	"sitehub/service/rbac/model/entity"

	"sitehub/service/rbac/rpc/internal/svc"
	"sitehub/service/rbac/rpc/pb/rbac"

	"github.com/zeromicro/go-zero/core/logx"
)

type PublicMenusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPublicMenusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PublicMenusLogic {
	return &PublicMenusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *PublicMenusLogic) PublicMenus(in *rbac.SystemRequest) (*rbac.PublicMenuPage, error) {
	items, err := l.svcCtx.Core.Systems.MenuTree(l.ctx, in.System, true)
	if err != nil {
		return nil, classify(err)
	}
	return &rbac.PublicMenuPage{Items: publicMenuItems(items)}, nil
}

func publicMenuItems(items []*entity.SysMenu) []*rbac.PublicMenu {
	result := make([]*rbac.PublicMenu, 0, len(items))
	for _, item := range items {
		result = append(result, &rbac.PublicMenu{Id: item.ID, Name: item.MenuName, Path: item.Path, Component: item.Component, Icon: item.Icon, IsVisible: item.IsVisible, SystemCode: item.SystemCode, Children: publicMenuItems(item.Children)})
	}
	return result
}
