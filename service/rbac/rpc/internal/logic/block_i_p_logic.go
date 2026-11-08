package logic

import (
	"context"
	"sitehub/service/rbac/rpc/internal/svc"
	pb "sitehub/service/rbac/rpc/pb/rbac"

	"github.com/zeromicro/go-zero/core/logx"
)

type BlockIPLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBlockIPLogic(ctx context.Context, s *svc.ServiceContext) *BlockIPLogic {
	return &BlockIPLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *BlockIPLogic) BlockIP(in *pb.BlacklistInput) (*pb.Empty, error) {
	actor, err := l.svcCtx.Require(l.ctx, "system:config:edit", "admin", true)
	if err != nil {
		return nil, err
	}
	_ = actor

	err = l.svcCtx.Core.Blacklist.AddToBlacklist(in.Ip, in.Reason)
	return done(err)
}
