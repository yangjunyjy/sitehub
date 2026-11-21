// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package releaseadmin

import (
	"context"
	"sitehub/service/product/api/internal/logic"
	domain "sitehub/service/product/domain"

	"sitehub/service/product/api/internal/svc"
	"sitehub/service/product/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateReleaseGroupLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateReleaseGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateReleaseGroupLogic {
	return &CreateReleaseGroupLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateReleaseGroupLogic) CreateReleaseGroup(req *types.ReleaseGroupInput) (resp *types.Response, err error) {
	if _, err := l.svcCtx.Require(l.ctx, "website:downloads:edit", "admin", true); err != nil {
		return nil, err
	}
	input := domain.ReleaseGroup{ID: req.ID, Name: req.Name, Description: req.Description, SortOrder: req.SortOrder, Revision: req.Revision}
	return logic.Response(l.svcCtx.Core.Releases.SaveGroup(l.ctx, input, true))
}
