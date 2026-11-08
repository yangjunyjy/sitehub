package logic

import (
	"context"
	"sitehub/common/protoconv"
	entity "sitehub/service/rbac/model/entity"
	"sitehub/service/rbac/rpc/internal/svc"
	pb "sitehub/service/rbac/rpc/pb/rbac"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CreateMenuLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateMenuLogic(ctx context.Context, s *svc.ServiceContext) *CreateMenuLogic {
	return &CreateMenuLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *CreateMenuLogic) CreateMenu(in *pb.CreateSysMenuRequest) (*pb.Empty, error) {
	actor, err := l.svcCtx.Require(l.ctx, "system:menu:edit", "admin", true)
	if err != nil {
		return nil, err
	}
	_ = actor
	if in.Status == nil || *in.Status > 1 || in.Type < 1 || in.Type > 3 {
		return nil, status.Error(codes.InvalidArgument, "menu requires type 1–3 and status 0 or 1")
	}
	var input entity.SysMenu
	if err := protoconv.Decode(in, &input); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	input.MenuName = in.Name
	input.MenuCode = in.Code
	input.MenuType = entity.MenuType(in.Type)
	input.Sort = int(in.Sort)
	input.Remark = in.Description
	err = l.svcCtx.Core.Menus.Create(&input)
	return done(err)
}
