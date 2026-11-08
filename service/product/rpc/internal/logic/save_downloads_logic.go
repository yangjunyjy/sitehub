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

type SaveDownloadsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSaveDownloadsLogic(ctx context.Context, s *svc.ServiceContext) *SaveDownloadsLogic {
	return &SaveDownloadsLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *SaveDownloadsLogic) SaveDownloads(in *pb.WebsiteDownloads) (*pb.Empty, error) {
	actor, err := l.svcCtx.Require(l.ctx, "website:downloads:edit", "admin", true)
	if err != nil {
		return nil, err
	}
	_ = actor
	var input contract.WebsiteDownloads
	if err := protoconv.Decode(in, &input); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	err = l.svcCtx.Core.Content.SaveDownloads(input)
	return done(err)
}
