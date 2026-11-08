package logic

import (
	"context"
	"sitehub/service/rbac/rpc/internal/svc"
	pb "sitehub/service/rbac/rpc/pb/rbac"

	"github.com/zeromicro/go-zero/core/logx"
)

type AuthorizeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAuthorizeLogic(ctx context.Context, s *svc.ServiceContext) *AuthorizeLogic {
	return &AuthorizeLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *AuthorizeLogic) Authorize(in *pb.AuthorizeRequest) (*pb.Principal, error) {
	return l.svcCtx.Authorize(l.ctx, in.Token, in.Permission, in.Audience, in.Secure, in.Totp)
}
