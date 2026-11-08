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

type CreateCategoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateCategoryLogic(ctx context.Context, s *svc.ServiceContext) *CreateCategoryLogic {
	return &CreateCategoryLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *CreateCategoryLogic) CreateCategory(in *pb.CategoryInput) (*pb.Empty, error) {
	actor, err := l.svcCtx.Require(l.ctx, "blog:category:add", "admin", true)
	if err != nil {
		return nil, err
	}
	_ = actor
	var input entity.Category
	if err := protoconv.Decode(in, &input); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	err = l.svcCtx.Core.Categories.CreateCategory(&input)
	return done(err)
}
