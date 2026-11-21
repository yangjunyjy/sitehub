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

type UpdateReleaseFileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateReleaseFileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateReleaseFileLogic {
	return &UpdateReleaseFileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateReleaseFileLogic) UpdateReleaseFile(req *types.ReleaseFileInput) (resp *types.Response, err error) {
	if _, err := l.svcCtx.Require(l.ctx, "website:downloads:edit", "admin", true); err != nil {
		return nil, err
	}
	input := domain.ReleaseFile{ID: req.ID, GroupID: req.GroupID, Product: req.Product, Edition: req.Edition, Version: req.Version, Platform: req.Platform, Arch: req.Arch, Format: req.Format, URL: req.URL, FileName: req.FileName, Description: req.Description, Icon: req.Icon, Checksum: req.Checksum, ReleaseDate: req.ReleaseDate, Size: req.Size, SortOrder: req.SortOrder, Revision: req.Revision}
	return logic.Response(l.svcCtx.Core.Releases.SaveFile(l.ctx, input, false))
}
