package logic

import (
	"context"
	"sitehub/common/rpcx"
	"sitehub/service/rbac/rpc/internal/svc"
	pb "sitehub/service/rbac/rpc/pb/rbac"

	"github.com/zeromicro/go-zero/core/logx"
)

type CurrentMemberLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCurrentMemberLogic(ctx context.Context, s *svc.ServiceContext) *CurrentMemberLogic {
	return &CurrentMemberLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *CurrentMemberLogic) CurrentMember(in *pb.Empty) (*pb.MemberProfile, error) {
	value, err := l.svcCtx.Core.Members.Profile(l.ctx, rpcx.Incoming(l.ctx, "authorization"))
	return reply[pb.MemberProfile](value, err)
}
