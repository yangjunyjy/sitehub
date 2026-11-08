package logic

import (
	"context"
	"sitehub/common/protoconv"
	"sitehub/common/rpcx"
	"sitehub/service/file/api/internal/svc"
	"sitehub/service/file/api/internal/types"
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
func (l *SaveStorageLogic) SaveStorage(req *types.SaveStorageRequest) (*types.Response, error) {
	in := &pb.WebsiteStorage{}
	if err := protoconv.Encode(req, in); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	out, err := l.svcCtx.RPC.SaveStorage(rpcx.Outgoing(l.ctx, l.svcCtx.Config.ServiceKey), in)
	if err != nil {
		return nil, err
	}
	value, err := protoconv.Value(out)
	if err != nil {
		return nil, err
	}
	return &types.Response{Code: 0, Message: "ok", Data: value}, nil
}
