package storage

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	logicshared "sitehub/service/blog/api/internal/logic"
	"sitehub/service/blog/api/internal/svc"
	"sitehub/service/blog/api/internal/types"
	contract "sitehub/service/blog/contract"
)

type SaveStorageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSaveStorageLogic(ctx context.Context, s *svc.ServiceContext) *SaveStorageLogic {
	return &SaveStorageLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *SaveStorageLogic) SaveStorage(req *types.StorageSettings) (*types.Response, error) {

	_, err := l.svcCtx.Require(l.ctx, "file:storage:edit", "admin", true)
	if err != nil {
		return nil, err
	}
	var input contract.WebsiteStorage
	if err := logicshared.DecodeRequest(req, &input); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	err = l.svcCtx.Core.StorageSettings.SaveStorage(input)
	return logicshared.Done(err)
}
