package logic

import (
	"context"
	"sitehub/common/protoconv"
	"sitehub/common/rpcx"
	"sitehub/service/file/api/internal/svc"
	"sitehub/service/file/api/internal/types"
	pb "sitehub/service/file/rpc/pb/file"

	"github.com/zeromicro/go-zero/core/logx"
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
	in := &pb.Empty{}
	out, err := l.svcCtx.RPC.GetStorage(rpcx.Outgoing(l.ctx, l.svcCtx.Config.ServiceKey), in)
	if err != nil {
		return nil, err
	}
	value, err := protoconv.Value(out)
	if err != nil {
		return nil, err
	}
	return &types.Response{Code: 0, Message: "ok", Data: value}, nil
}
