package settings

import (
	"context"
	logicshared "sitehub/service/blog/api/internal/logic"
	"strings"
	"unicode/utf8"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sitehub/service/blog/api/internal/svc"
	"sitehub/service/blog/api/internal/types"
)

type SaveHomeSettingsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSaveHomeSettingsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveHomeSettingsLogic {
	return &SaveHomeSettingsLogic{ctx, svcCtx}
}

func (logic *SaveHomeSettingsLogic) SaveHomeSettings(req *types.HomeSettingsRequest) (*types.Response, error) {
	if _, err := logic.svcCtx.Require(logic.ctx, "system:settings:edit", "admin", true); err != nil {
		return nil, err
	}
	verse := strings.TrimSpace(req.HomeVerse)
	if verse == "" || utf8.RuneCountInString(verse) > 120 {
		return nil, status.Error(codes.InvalidArgument, "首页短句需为 1 至 120 个字符")
	}
	return logicshared.Done(logic.svcCtx.Core.Community.SaveHome(logic.ctx, req.HomeAnimation, verse))
}
