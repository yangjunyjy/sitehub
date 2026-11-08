package logic

import (
	"context"
	"sitehub/common/rpcx"
	"sitehub/service/blog/rpc/internal/svc"
	pb "sitehub/service/blog/rpc/pb/blog"

	"github.com/zeromicro/go-zero/core/logx"
)

type RecordViewLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRecordViewLogic(ctx context.Context, s *svc.ServiceContext) *RecordViewLogic {
	return &RecordViewLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *RecordViewLogic) RecordView(in *pb.ViewRequest) (*pb.Empty, error) {
	if _, err := l.svcCtx.Core.Articles.PublicDetail(uint64(in.Id)); err != nil {
		return nil, rpcx.Error(err)
	}
	return done(l.svcCtx.Core.Articles.RecordViewWithInfo(uint(in.Id), 0, rpcx.Incoming(l.ctx, "x-client-ip"), rpcx.Incoming(l.ctx, "x-client-agent")))
}
