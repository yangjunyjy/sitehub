package logic

import (
	"context"
	"sitehub/common/rpcx"
	"sitehub/common/streamio"
	domain "sitehub/service/blog/domain"
	"sitehub/service/blog/rpc/internal/svc"
	pb "sitehub/service/blog/rpc/pb/blog"
)

type ReadAssetLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReadAssetLogic(ctx context.Context, s *svc.ServiceContext) *ReadAssetLogic {
	return &ReadAssetLogic{ctx, s}
}
func (l *ReadAssetLogic) ReadAsset(in *pb.AssetRequest, stream pb.Blog_ReadAssetServer) error {
	actor := domain.ArticleActor{}
	if rpcx.Incoming(stream.Context(), "authorization") != "" {
		principal, err := l.svcCtx.Require(stream.Context(), "", "admin", false)
		if err != nil {
			return err
		}
		actor = articleActor(principal)
	}
	value, err := l.svcCtx.Core.Assets.ReadArticleAsset(stream.Context(), actor, in.Id)
	if err != nil {
		return classify(err)
	}
	defer value.Body.Close()
	if err = stream.Send(&pb.StreamChunk{Name: value.Name, Size: value.Size, Mime: value.MIME}); err != nil {
		return err
	}
	return streamio.Copy(stream.Context(), value.Body, func(data []byte) error { return stream.Send(&pb.StreamChunk{Data: data}) })
}
