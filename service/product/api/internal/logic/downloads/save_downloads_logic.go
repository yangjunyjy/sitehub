package downloads

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"sitehub/service/product/api/internal/logic"
	"sitehub/service/product/api/internal/svc"
	"sitehub/service/product/api/internal/types"
	"sitehub/service/product/contract"
)

type SaveDownloadsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSaveDownloadsLogic(ctx context.Context, s *svc.ServiceContext) *SaveDownloadsLogic {
	return &SaveDownloadsLogic{ctx, s, logx.WithContext(ctx)}
}

func (l *SaveDownloadsLogic) SaveDownloads(req *types.SaveDownloadsRequest) (*types.Response, error) {
	_, err := l.svcCtx.Require(l.ctx, "website:downloads:edit", "admin", true)
	if err != nil {
		return nil, err
	}
	var input contract.WebsiteDownloads
	if err := logic.CopyManifest(req, &input); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	err = l.svcCtx.Core.Content.SaveDownloads(input)
	return logic.Response(struct{}{}, err)
}
