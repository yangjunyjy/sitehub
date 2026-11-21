package content

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"sitehub/service/product/api/internal/logic"
	"sitehub/service/product/api/internal/svc"
	"sitehub/service/product/api/internal/types"
)

type ContentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewContentLogic(ctx context.Context, s *svc.ServiceContext) *ContentLogic {
	return &ContentLogic{ctx, s, logx.WithContext(ctx)}
}

func (l *ContentLogic) Content() (*types.Response, error) {
	value, err := l.svcCtx.Core.Content.GetContent()
	return logic.Response(value, err)
}
