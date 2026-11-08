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

type UpdateCategoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateCategoryLogic(ctx context.Context, s *svc.ServiceContext) *UpdateCategoryLogic {
	return &UpdateCategoryLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *UpdateCategoryLogic) UpdateCategory(in *pb.CategoryInput) (*pb.Empty, error) {
	actor, err := l.svcCtx.Require(l.ctx, "blog:category:edit", "admin", true)
	if err != nil {
		return nil, err
	}
	_ = actor
	var input entity.Category
	if err := protoconv.Decode(in, &input); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	err = l.svcCtx.Core.Categories.UpdateCategory(uint(in.Id), &input)
	return done(err)
}
