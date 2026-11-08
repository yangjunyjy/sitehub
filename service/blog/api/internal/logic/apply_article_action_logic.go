package logic

import (
	"context"
	"sitehub/common/protoconv"
	"sitehub/common/rpcx"
	"sitehub/service/blog/api/internal/svc"
	"sitehub/service/blog/api/internal/types"
	pb "sitehub/service/blog/rpc/pb/blog"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ApplyArticleActionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewApplyArticleActionLogic(ctx context.Context, s *svc.ServiceContext) *ApplyArticleActionLogic {
	return &ApplyArticleActionLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *ApplyArticleActionLogic) ApplyArticleAction(req *types.ApplyArticleActionRequest) (*types.Response, error) {
	in := &pb.ArticleAction{}
	if err := protoconv.Encode(req, in); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	out, err := l.svcCtx.RPC.ApplyArticleAction(rpcx.Outgoing(l.ctx, l.svcCtx.Config.ServiceKey), in)
	if err != nil {
		return nil, err
	}
	value, err := protoconv.Value(out)
	if err != nil {
		return nil, err
	}
	return &types.Response{Code: 0, Message: "ok", Data: value}, nil
}
