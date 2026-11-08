package logic

import (
	"context"
	"sitehub/service/rbac/rpc/internal/svc"
	pb "sitehub/service/rbac/rpc/pb/rbac"

	"github.com/zeromicro/go-zero/core/logx"
)

type UnblockIPLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUnblockIPLogic(ctx context.Context, s *svc.ServiceContext) *UnblockIPLogic {
	return &UnblockIPLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *UnblockIPLogic) UnblockIP(in *pb.BlacklistInput) (*pb.Empty, error) {
	actor, err := l.svcCtx.Require(l.ctx, "system:config:edit", "admin", true)
	if err != nil {
		return nil, err
	}
	_ = actor

	err = l.svcCtx.Core.Blacklist.RemoveFromBlacklist(uint(in.Id))
	return done(err)
}
