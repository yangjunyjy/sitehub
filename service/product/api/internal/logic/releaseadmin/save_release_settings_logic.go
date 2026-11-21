// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package releaseadmin

import (
	"context"
	"sitehub/service/product/api/internal/logic"
	"sitehub/service/product/contract"
	domain "sitehub/service/product/domain"

	"sitehub/service/product/api/internal/svc"
	"sitehub/service/product/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SaveReleaseSettingsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSaveReleaseSettingsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveReleaseSettingsLogic {
	return &SaveReleaseSettingsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SaveReleaseSettingsLogic) SaveReleaseSettings(req *types.ReleaseSettingsInput) (resp *types.Response, err error) {
	if _, err := l.svcCtx.Require(l.ctx, "website:downloads:edit", "admin", true); err != nil {
		return nil, err
	}
	input := domain.ReleaseSettings{Title: req.Title, Description: req.Description, Version: req.Version, ReleaseDate: req.ReleaseDate, SupportArchitectures: req.SupportArchitectures, SupportTitle: req.SupportTitle, SupportDescription: req.SupportDescription, Revision: req.Revision}
	for _, v := range req.InstallOptions {
		input.InstallOptions = append(input.InstallOptions, contract.WebsiteInstallOption{ID: v.ID, Title: v.Title, Badge: v.Badge, Description: v.Description, Command: v.Command})
	}
	for _, v := range req.Requirements {
		input.Requirements = append(input.Requirements, contract.WebsiteRequirement{Label: v.Label, Value: v.Value})
	}
	return logic.Response(l.svcCtx.Core.Releases.SaveSettings(l.ctx, input))
}
