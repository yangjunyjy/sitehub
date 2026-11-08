package logic

import (
	"context"
	"sitehub/common/protoconv"
	"sitehub/common/rpcx"
	"sitehub/service/product/api/internal/svc"
	"sitehub/service/product/api/internal/types"
	pb "sitehub/service/product/rpc/pb/product"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ChangeDocModeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewChangeDocModeLogic(ctx context.Context, s *svc.ServiceContext) *ChangeDocModeLogic {
	return &ChangeDocModeLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *ChangeDocModeLogic) ChangeDocMode(req *types.ChangeDocModeRequest) (*types.Response, error) {
	in := &pb.DocModeInput{}
	if err := protoconv.Encode(req, in); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	out, err := l.svcCtx.RPC.ChangeDocMode(rpcx.Outgoing(l.ctx, l.svcCtx.Config.ServiceKey), in)
	if err != nil {
		return nil, err
	}
	value, err := protoconv.Value(out)
	if err != nil {
		return nil, err
	}
	return &types.Response{Code: 0, Message: "ok", Data: value}, nil
}
