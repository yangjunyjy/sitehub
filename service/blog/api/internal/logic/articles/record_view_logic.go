package articles

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	"sitehub/common/rpcx"
	logicshared "sitehub/service/blog/api/internal/logic"
	"sitehub/service/blog/api/internal/svc"
	"sitehub/service/blog/api/internal/types"
)

type RecordViewLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRecordViewLogic(ctx context.Context, s *svc.ServiceContext) *RecordViewLogic {
	return &RecordViewLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *RecordViewLogic) RecordView(req *types.RecordViewRequest) (*types.Response, error) {

	if _, err := l.svcCtx.Core.Articles.PublicDetail(uint64(req.ID)); err != nil {
		return nil, rpcx.Error(err)
	}
	return logicshared.Done(l.svcCtx.Core.Articles.RecordViewWithInfo(uint(req.ID), 0, rpcx.HTTPValue(l.ctx, "x-client-ip"), rpcx.HTTPValue(l.ctx, "x-client-agent")))
}
