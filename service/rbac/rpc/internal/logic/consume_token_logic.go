package logic

import (
	"context"
	"errors"
	"sitehub/common/pkg/token"
	"sitehub/common/rpcx"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"sitehub/service/rbac/rpc/internal/svc"
	"sitehub/service/rbac/rpc/pb/rbac"

	"github.com/zeromicro/go-zero/core/logx"
)

type ConsumeTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewConsumeTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ConsumeTokenLogic {
	return &ConsumeTokenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ConsumeTokenLogic) ConsumeToken(in *rbac.ConsumeTokenRequest) (*rbac.Empty, error) {
	actor, err := l.svcCtx.Require(l.ctx, "", "admin", false)
	if err != nil {
		return nil, err
	}
	value, err := l.svcCtx.Core.Tokens.Verify(l.ctx, in.Token)
	if errors.Is(err, token.ErrTokenNotFound) || errors.Is(err, token.ErrTokenExpired) {
		return nil, status.Error(codes.Unauthenticated, "token expired or already consumed")
	}
	if err != nil {
		return nil, rpcx.Error(err)
	}
	if value.UserID != actor.Id {
		return nil, status.Error(codes.PermissionDenied, "token owner mismatch")
	}
	return &rbac.Empty{}, nil
}
