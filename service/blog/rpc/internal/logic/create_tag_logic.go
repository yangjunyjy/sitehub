package logic

import (
	"context"
	"sitehub/common/protoconv"
	entity "sitehub/service/blog/model/entity"
	"sitehub/service/blog/rpc/internal/svc"
	pb "sitehub/service/blog/rpc/pb/blog"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CreateTagLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateTagLogic(ctx context.Context, s *svc.ServiceContext) *CreateTagLogic {
	return &CreateTagLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *CreateTagLogic) CreateTag(in *pb.CreateTagRequest) (*pb.Empty, error) {
	actor, err := l.svcCtx.Require(l.ctx, "blog:tag:add", "admin", true)
	if err != nil {
		return nil, err
	}
	_ = actor
	var input entity.Tag
	if err := protoconv.Decode(in, &input); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	err = l.svcCtx.Core.Tags.CreateTag(&input)
	return done(err)
}
