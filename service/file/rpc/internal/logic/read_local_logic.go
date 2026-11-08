package logic

import (
	"context"
	"io"
	"mime"
	"sitehub/common/rpcx"
	"sitehub/common/streamio"
	"sitehub/service/file/rpc/internal/svc"
	pb "sitehub/service/file/rpc/pb/file"
	"os"
	"path/filepath"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ReadLocalLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReadLocalLogic(ctx context.Context, s *svc.ServiceContext) *ReadLocalLogic {
	return &ReadLocalLogic{ctx, s}
}
func (l *ReadLocalLogic) ReadLocal(in *pb.LocationRequest, stream pb.File_ReadLocalServer) error {
	name, err := l.svcCtx.Core.Uploads.ServeLocalPath(in.Location)
	if err != nil {
		return classify(err)
	}
	file, err := os.Open(name)
	if err != nil {
		return rpcx.Error(err)
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return rpcx.Error(err)
	}
	if in.Offset < 0 || in.Offset > info.Size() || in.Length < 0 {
		return status.Error(codes.InvalidArgument, "invalid range")
	}
	if _, err = file.Seek(in.Offset, io.SeekStart); err != nil {
		return rpcx.Error(err)
	}
	if err = stream.Send(&pb.StreamChunk{Name: info.Name(), Size: info.Size(), Mime: mime.TypeByExtension(filepath.Ext(info.Name()))}); err != nil {
		return err
	}
	var reader io.Reader = file
	if in.Length > 0 {
		reader = io.LimitReader(file, in.Length)
	}
	return streamio.Copy(stream.Context(), reader, func(data []byte) error { return stream.Send(&pb.StreamChunk{Data: data}) })
}
