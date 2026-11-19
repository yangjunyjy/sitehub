package logic

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	"sitehub/service/blog/api/internal/payload"
	"sitehub/service/blog/api/internal/svc"
	"sitehub/service/blog/api/internal/types"
)

type GetStorageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetStorageLogic(ctx context.Context, s *svc.ServiceContext) *GetStorageLogic {
	return &GetStorageLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *GetStorageLogic) GetStorage() (*types.Response, error) {

	_, err := l.svcCtx.Require(l.ctx, "file:storage:list", "admin", false)
	if err != nil {
		return nil, err
	}
	value, err := l.svcCtx.Core.StorageSettings.GetStorage()
	value.SecretKey = ""
	return ReplyAs[payload.WebsiteStorage](value, err)
}
