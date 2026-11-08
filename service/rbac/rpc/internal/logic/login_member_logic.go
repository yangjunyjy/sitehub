package logic

import (
	"context"
	"sitehub/common/protoconv"
	"sitehub/common/rpcx"
	domain "sitehub/service/rbac/domain"
	"sitehub/service/rbac/rpc/internal/svc"
	pb "sitehub/service/rbac/rpc/pb/rbac"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginMemberLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginMemberLogic(ctx context.Context, s *svc.ServiceContext) *LoginMemberLogic {
	return &LoginMemberLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *LoginMemberLogic) LoginMember(in *pb.LoginMemberInput) (*pb.MemberSessionReply, error) {
	var input domain.LoginMemberInput
	if err := protoconv.Decode(in, &input); err != nil {
		return nil, err
	}
	user, token, err := l.svcCtx.Core.Members.Login(l.ctx, input, rpcx.Incoming(l.ctx, "x-client-ip"))
	return reply[pb.MemberSessionReply](map[string]any{"user": user, "token": token, "expiresAt": time.Now().Add(7 * 24 * time.Hour).Unix()}, memberError(err))
}
