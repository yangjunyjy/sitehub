package logic

import (
	"context"
	"sitehub/service/rbac/rpc/internal/svc"
	pb "sitehub/service/rbac/rpc/pb/rbac"

	"github.com/zeromicro/go-zero/core/logx"
)

type CurrentPermissionsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCurrentPermissionsLogic(ctx context.Context, s *svc.ServiceContext) *CurrentPermissionsLogic {
	return &CurrentPermissionsLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *CurrentPermissionsLogic) CurrentPermissions(in *pb.Empty) (*pb.StringList, error) {
	actor, err := l.svcCtx.Require(l.ctx, "", "admin", false)
	if err != nil {
		return nil, err
	}
	_ = actor
	return &pb.StringList{Items: actor.Permissions}, nil
}
