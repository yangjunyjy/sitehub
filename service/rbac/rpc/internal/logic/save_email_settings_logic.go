package logic

import (
	"context"
	"sitehub/common/protoconv"
	request "sitehub/service/rbac/dto/request"
	"sitehub/service/rbac/rpc/internal/svc"
	pb "sitehub/service/rbac/rpc/pb/rbac"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SaveEmailSettingsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSaveEmailSettingsLogic(ctx context.Context, s *svc.ServiceContext) *SaveEmailSettingsLogic {
	return &SaveEmailSettingsLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *SaveEmailSettingsLogic) SaveEmailSettings(in *pb.UpdateEmailSettingsRequest) (*pb.EmailSettingsView, error) {
	actor, err := l.svcCtx.Require(l.ctx, "system:config:edit", "admin", true)
	if err != nil {
		return nil, err
	}
	_ = actor
	var input request.UpdateEmailSettingsRequest
	if err := protoconv.Decode(in, &input); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	value, err := l.svcCtx.Core.Email.Save(input)
	return reply[pb.EmailSettingsView](value, err)
}
