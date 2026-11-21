package documents

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"sitehub/service/product/api/internal/logic"
	"sitehub/service/product/api/internal/svc"
	"sitehub/service/product/api/internal/types"
	"sitehub/service/product/domain"
)

type SaveDocumentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSaveDocumentLogic(ctx context.Context, s *svc.ServiceContext) *SaveDocumentLogic {
	return &SaveDocumentLogic{ctx, s, logx.WithContext(ctx)}
}

func (l *SaveDocumentLogic) SaveDocument(req *types.SaveDocumentRequest) (*types.Response, error) {
	actor, err := l.svcCtx.Require(l.ctx, "website:docs:edit", "admin", true)
	if err != nil {
		return nil, err
	}
	input := domain.DocumentInput{
		Version:  req.Version,
		ID:       req.ID,
		Slug:     req.Slug,
		Title:    req.Title,
		Summary:  req.Summary,
		Content:  req.Content,
		GroupID:  req.GroupID,
		Revision: req.Revision,
	}
	value, err := l.svcCtx.Core.Content.SaveDocument(input, actor.Id)
	return logic.Response(value, err)
}
