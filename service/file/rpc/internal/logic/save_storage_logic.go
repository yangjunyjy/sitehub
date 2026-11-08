package logic

import (
	"context"
	"sitehub/common/protoconv"
	contract "sitehub/service/file/contract"
	"sitehub/service/file/rpc/internal/svc"
	pb "sitehub/service/file/rpc/pb/file"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SaveStorageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSaveStorageLogic(ctx context.Context, s *svc.ServiceContext) *SaveStorageLogic {
	return &SaveStorageLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *SaveStorageLogic) SaveStorage(in *pb.WebsiteStorage) (*pb.Empty, error) {
	actor, err := l.svcCtx.Require(l.ctx, "website:storage:edit", "admin", true)
	if err != nil {
		return nil, err
	}
	_ = actor
	var input contract.WebsiteStorage
	if err := protoconv.Decode(in, &input); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	err = l.svcCtx.Core.Settings.SaveStorage(input)
	return done(err)
}
