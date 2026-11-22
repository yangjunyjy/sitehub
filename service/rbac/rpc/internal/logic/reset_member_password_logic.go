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

type ResetMemberPasswordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewResetMemberPasswordLogic(ctx context.Context, s *svc.ServiceContext) *ResetMemberPasswordLogic {
	return &ResetMemberPasswordLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *ResetMemberPasswordLogic) ResetMemberPassword(in *pb.ResetMemberPasswordInput) (*pb.Empty, error) {
	var input domain.ResetMemberPasswordInput
	if err := protoconv.Decode(in, &input); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	return done(l.svcCtx.Core.Members.ResetPassword(l.ctx, input, rpcx.Incoming(l.ctx, "x-client-ip")))
}
