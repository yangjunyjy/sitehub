package logic

import (
	"context"
	"sitehub/common/rpcx"
	"sitehub/service/rbac/rpc/internal/svc"
	pb "sitehub/service/rbac/rpc/pb/rbac"

	"github.com/zeromicro/go-zero/core/logx"
)

type OnceTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewOnceTokenLogic(ctx context.Context, s *svc.ServiceContext) *OnceTokenLogic {
	return &OnceTokenLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *OnceTokenLogic) OnceToken(in *pb.Empty) (*pb.TokenReply, error) {
	actor, err := l.svcCtx.Require(l.ctx, "", "admin", true)
	if err != nil {
		return nil, err
	}
	_ = actor
	value, err := l.svcCtx.Core.Tokens.Generate(l.ctx, actor.Id, actor.Name, nil)
	if err != nil {
		return nil, rpcx.Error(err)
	}
	return &pb.TokenReply{Token: value.TokenID, ExpiresAt: value.ExpiresAt}, nil
}
