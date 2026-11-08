package logic

import (
	"context"
	"sitehub/common/protoconv"
	"sitehub/common/rpcx"
	"sitehub/common/streamio"
	business "sitehub/service/file/business"
	"sitehub/service/file/rpc/internal/svc"
	pb "sitehub/service/file/rpc/pb/file"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UploadLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUploadLogic(ctx context.Context, s *svc.ServiceContext) *UploadLogic {
	return &UploadLogic{ctx, s}
}
func (l *UploadLogic) Upload(stream pb.File_UploadServer) error {
	if _, err := l.svcCtx.Require(stream.Context(), "website:storage:edit", "admin", true); err != nil {
		return err
	}
	file, meta, err := streamio.Receive(stream.Context(), stream.Recv, 2<<30)
	if err != nil {
		return rpcx.Error(err)
	}
	defer streamio.Remove(file)
	var value *business.AttachmentInfo
	switch meta.Kind {
	case "image":
		var location string
		location, err = l.svcCtx.Core.Uploads.UploadImage(meta.Name, meta.Size, file)
		value = &business.AttachmentInfo{URL: location, Name: meta.Name, Size: meta.Size}
	case "attachment":
		value, err = l.svcCtx.Core.Uploads.UploadAttachment(meta.Name, meta.Size, file)
	case "release":
		value, err = l.svcCtx.Core.Uploads.UploadRelease(meta.Name, meta.Size, file)
	default:
		return status.Error(codes.InvalidArgument, "unknown upload kind")
	}
	if err != nil {
		return classify(err)
	}
	out := &pb.FileInfo{}
	if err = protoconv.Encode(value, out); err != nil {
		return rpcx.Error(err)
	}
	return stream.SendAndClose(out)
}
