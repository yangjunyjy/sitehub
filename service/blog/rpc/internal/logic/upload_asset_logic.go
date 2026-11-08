package logic

import (
	"context"
	"sitehub/common/protoconv"
	"sitehub/common/rpcx"
	"sitehub/common/streamio"
	"sitehub/service/blog/rpc/internal/svc"
	pb "sitehub/service/blog/rpc/pb/blog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UploadAssetLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUploadAssetLogic(ctx context.Context, s *svc.ServiceContext) *UploadAssetLogic {
	return &UploadAssetLogic{ctx, s}
}
func (l *UploadAssetLogic) UploadAsset(stream pb.Blog_UploadAssetServer) error {
	principal, err := l.svcCtx.Require(stream.Context(), "", "admin", true)
	if err != nil {
		return err
	}
	actor := articleActor(principal)
	if !actor.CanCreate && !actor.CanUpdate {
		return status.Error(codes.PermissionDenied, "permission denied")
	}
	file, meta, err := streamio.Receive(stream.Context(), stream.Recv, 50<<20)
	if err != nil {
		return rpcx.Error(err)
	}
	defer streamio.Remove(file)
	value, err := l.svcCtx.Core.Assets.UploadArticleAsset(stream.Context(), actor, meta.Name, meta.Size, file)
	if err != nil {
		return classify(err)
	}
	out := &pb.AssetInfo{}
	if err = protoconv.Encode(value, out); err != nil {
		return rpcx.Error(err)
	}
	return stream.SendAndClose(out)
}
