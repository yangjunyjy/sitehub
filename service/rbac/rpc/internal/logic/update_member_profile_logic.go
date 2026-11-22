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

type UpdateMemberProfileLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateMemberProfileLogic(ctx context.Context, s *svc.ServiceContext) *UpdateMemberProfileLogic {
	return &UpdateMemberProfileLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *UpdateMemberProfileLogic) UpdateMemberProfile(in *pb.UpdateMemberProfileInput) (*pb.MemberProfile, error) {
	var input domain.UpdateMemberProfileInput
	if err := protoconv.Decode(in, &input); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	value, err := l.svcCtx.Core.Members.UpdateProfile(l.ctx, rpcx.Incoming(l.ctx, "authorization"), input)
	return reply[pb.MemberProfile](value, err)
}
