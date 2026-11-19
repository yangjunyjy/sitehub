package logic

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sitehub/service/blog/api/internal/svc"
	"sitehub/service/blog/api/internal/types"
	entity "sitehub/service/blog/model/entity"
)

type CreateTagLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateTagLogic(ctx context.Context, s *svc.ServiceContext) *CreateTagLogic {
	return &CreateTagLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *CreateTagLogic) CreateTag(req *types.CreateTagRequest) (*types.Response, error) {

	_, err := l.svcCtx.Require(l.ctx, "blog:tag:create", "admin", true)
	if err != nil {
		return nil, err
	}
	var input entity.Tag
	if err := decodeRequest(req, &input); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	err = l.svcCtx.Core.Tags.CreateTag(&input)
	return done(err)
}
