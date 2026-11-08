package logic

import (
	"context"
	"sitehub/common/protoconv"
	"sitehub/common/rpcx"
	"sitehub/service/rbac/api/internal/svc"
	"sitehub/service/rbac/api/internal/types"
	pb "sitehub/service/rbac/rpc/pb/rbac"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetEmailSettingsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetEmailSettingsLogic(ctx context.Context, s *svc.ServiceContext) *GetEmailSettingsLogic {
	return &GetEmailSettingsLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *GetEmailSettingsLogic) GetEmailSettings() (*types.Response, error) {
	in := &pb.Empty{}
	out, err := l.svcCtx.RPC.GetEmailSettings(rpcx.Outgoing(l.ctx, l.svcCtx.Config.ServiceKey), in)
	if err != nil {
		return nil, err
	}
	value, err := protoconv.Value(out)
	if err != nil {
		return nil, err
	}
	return &types.Response{Code: 0, Message: "ok", Data: value}, nil
}
