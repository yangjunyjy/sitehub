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

type ResolveDownloadLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewResolveDownloadLogic(ctx context.Context, s *svc.ServiceContext) *ResolveDownloadLogic {
	return &ResolveDownloadLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *ResolveDownloadLogic) ResolveDownload(req *types.ResolveDownloadRequest) (*types.Response, error) {
	in := &pb.DownloadRequest{}
	if err := protoconv.Encode(req, in); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	out, err := l.svcCtx.RPC.ResolveDownload(rpcx.Outgoing(l.ctx, l.svcCtx.Config.ServiceKey), in)
	if err != nil {
		return nil, err
	}
	value, err := protoconv.Value(out)
	if err != nil {
		return nil, err
	}
	return &types.Response{Code: 0, Message: "ok", Data: value}, nil
}
