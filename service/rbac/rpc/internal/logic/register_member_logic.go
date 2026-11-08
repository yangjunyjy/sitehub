package logic

import (
	"context"
	"sitehub/common/protoconv"
	"sitehub/common/rpcx"
	domain "sitehub/service/rbac/domain"
	"sitehub/service/rbac/rpc/internal/svc"
	pb "sitehub/service/rbac/rpc/pb/rbac"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RegisterMemberLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRegisterMemberLogic(ctx context.Context, s *svc.ServiceContext) *RegisterMemberLogic {
	return &RegisterMemberLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *RegisterMemberLogic) RegisterMember(in *pb.RegisterMemberInput) (*pb.MemberProfile, error) {
	var input domain.RegisterMemberInput
	if err := protoconv.Decode(in, &input); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	value, err := l.svcCtx.Core.Members.Register(l.ctx, input, rpcx.Incoming(l.ctx, "x-client-ip"))
	return reply[pb.MemberProfile](value, err)
}
