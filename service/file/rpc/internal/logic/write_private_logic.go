package logic

import (
	"context"
	"sitehub/common/rpcx"
	"sitehub/common/streamio"
	"sitehub/service/file/rpc/internal/svc"
	pb "sitehub/service/file/rpc/pb/file"
)

type WritePrivateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWritePrivateLogic(ctx context.Context, s *svc.ServiceContext) *WritePrivateLogic {
	return &WritePrivateLogic{ctx, s}
}
func (l *WritePrivateLogic) WritePrivate(stream pb.File_WritePrivateServer) error {
	file, meta, err := streamio.Receive(stream.Context(), stream.Recv, 50<<20)
	if err != nil {
		return rpcx.Error(err)
	}
	defer streamio.Remove(file)
	if err = l.svcCtx.Core.Private.WritePrivate(stream.Context(), meta.Snapshot, file, meta.Mime); err != nil {
		return classify(err)
	}
	return stream.SendAndClose(&pb.Empty{})
}
