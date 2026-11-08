package logic

import (
	"context"
	"sitehub/common/protoconv"
	contract "sitehub/service/product/contract"
	"sitehub/service/product/rpc/internal/svc"
	pb "sitehub/service/product/rpc/pb/product"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SaveContentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSaveContentLogic(ctx context.Context, s *svc.ServiceContext) *SaveContentLogic {
	return &SaveContentLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *SaveContentLogic) SaveContent(in *pb.WebsiteContent) (*pb.Empty, error) {
	actor, err := l.svcCtx.Require(l.ctx, "website:content:edit", "admin", true)
	if err != nil {
		return nil, err
	}
	_ = actor
	var input contract.WebsiteContent
	if err := protoconv.Decode(in, &input); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	err = l.svcCtx.Core.Content.SaveContent(input)
	return done(err)
}
