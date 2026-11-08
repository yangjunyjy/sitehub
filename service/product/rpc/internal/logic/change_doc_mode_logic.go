package logic

import (
	"context"
	"sitehub/common/protoconv"
	domain "sitehub/service/product/domain"
	"sitehub/service/product/rpc/internal/svc"
	pb "sitehub/service/product/rpc/pb/product"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ChangeDocModeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewChangeDocModeLogic(ctx context.Context, s *svc.ServiceContext) *ChangeDocModeLogic {
	return &ChangeDocModeLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *ChangeDocModeLogic) ChangeDocMode(in *pb.DocModeInput) (*pb.Empty, error) {
	actor, err := l.svcCtx.Require(l.ctx, "website:docs:edit", "admin", true)
	if err != nil {
		return nil, err
	}
	_ = actor
	var input domain.DocModeInput
	if err := protoconv.Decode(in, &input); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	err = l.svcCtx.Core.Content.ChangeDocMode(input)
	return done(err)
}
