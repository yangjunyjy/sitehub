package logic

import (
	"context"
	"sitehub/common/streamio"
	"sitehub/service/file/rpc/internal/svc"
	pb "sitehub/service/file/rpc/pb/file"
)

type ReadPrivateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReadPrivateLogic(ctx context.Context, s *svc.ServiceContext) *ReadPrivateLogic {
	return &ReadPrivateLogic{ctx, s}
}
func (l *ReadPrivateLogic) ReadPrivate(in *pb.SnapshotRequest, stream pb.File_ReadPrivateServer) error {
	body, err := l.svcCtx.Core.Private.ReadPrivate(stream.Context(), in.Snapshot)
	if err != nil {
		return classify(err)
	}
	defer body.Close()
	return streamio.Copy(stream.Context(), body, func(data []byte) error { return stream.Send(&pb.StreamChunk{Data: data}) })
}
