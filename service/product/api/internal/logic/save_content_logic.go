package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"sitehub/service/product/api/internal/svc"
	"sitehub/service/product/api/internal/types"
	"sitehub/service/product/contract"
)

type SaveContentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSaveContentLogic(ctx context.Context, s *svc.ServiceContext) *SaveContentLogic {
	return &SaveContentLogic{ctx, s, logx.WithContext(ctx)}
}

func (l *SaveContentLogic) SaveContent(req *types.SaveContentRequest) (*types.Response, error) {
	_, err := l.svcCtx.Require(l.ctx, "website:content:edit", "admin", true)
	if err != nil {
		return nil, err
	}
	var input contract.WebsiteContent
	if err := copyManifest(req, &input); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	err = l.svcCtx.Core.Content.SaveContent(input)
	return response(struct{}{}, err)
}
