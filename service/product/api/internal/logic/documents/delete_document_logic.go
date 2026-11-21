package documents

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"sitehub/service/product/api/internal/logic"
	"sitehub/service/product/api/internal/svc"
	"sitehub/service/product/api/internal/types"
	"sitehub/service/product/domain"
)

type DeleteDocumentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteDocumentLogic(ctx context.Context, s *svc.ServiceContext) *DeleteDocumentLogic {
	return &DeleteDocumentLogic{ctx, s, logx.WithContext(ctx)}
}

func (l *DeleteDocumentLogic) DeleteDocument(req *types.DeleteDocumentRequest) (*types.Response, error) {
	actor, err := l.svcCtx.Require(l.ctx, "website:docs:edit", "admin", true)
	if err != nil {
		return nil, err
	}
	input := domain.DocumentDeleteInput{
		Version:  req.Version,
		ID:       req.ID,
		Revision: req.Revision,
	}

	err = l.svcCtx.Core.Content.DeleteDocument(input, actor.Id)
	return logic.Response(struct{}{}, err)
}
