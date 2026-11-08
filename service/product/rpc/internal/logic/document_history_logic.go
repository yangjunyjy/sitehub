package logic

import (
	"context"
	"sitehub/service/product/rpc/internal/svc"
	pb "sitehub/service/product/rpc/pb/product"

	"github.com/zeromicro/go-zero/core/logx"
)

type DocumentHistoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDocumentHistoryLogic(ctx context.Context, s *svc.ServiceContext) *DocumentHistoryLogic {
	return &DocumentHistoryLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *DocumentHistoryLogic) DocumentHistory(in *pb.DocumentQuery) (*pb.DocumentHistoryReply, error) {
	actor, err := l.svcCtx.Require(l.ctx, "website:docs:edit", "admin", false)
	if err != nil {
		return nil, err
	}
	_ = actor
	value, err := l.svcCtx.Core.Content.DocumentHistory(in.Id, in.Version, in.HistoryId)
	if in.HistoryId > 0 {
		return reply[pb.DocumentHistoryReply](map[string]any{"detail": value}, err)
	}
	return reply[pb.DocumentHistoryReply](map[string]any{"items": value}, err)
}
