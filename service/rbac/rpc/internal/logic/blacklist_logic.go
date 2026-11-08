package logic

import (
	"context"
	"sitehub/common/rpcx"
	"sitehub/service/rbac/rpc/internal/svc"
	pb "sitehub/service/rbac/rpc/pb/rbac"

	"github.com/zeromicro/go-zero/core/logx"
)

type BlacklistLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBlacklistLogic(ctx context.Context, s *svc.ServiceContext) *BlacklistLogic {
	return &BlacklistLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *BlacklistLogic) Blacklist(in *pb.Empty) (*pb.StringList, error) {
	actor, err := l.svcCtx.Require(l.ctx, "system:config:list", "admin", false)
	if err != nil {
		return nil, err
	}
	_ = actor
	items, err := l.svcCtx.Core.Blacklist.BlockedIPs()
	return &pb.StringList{Items: items}, rpcx.Error(err)
}
