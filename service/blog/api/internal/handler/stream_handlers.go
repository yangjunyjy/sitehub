package handler

import (
	"sitehub/common/protoconv"
	"sitehub/common/rpcx"
	"sitehub/common/streamio"
	"sitehub/service/blog/api/internal/svc"
	pb "sitehub/service/blog/rpc/pb/blog"
	"net/http"
	"time"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func RegisterStreamHandlers(server *rest.Server, s *svc.ServiceContext) {
	server.AddRoutes([]rest.Route{{Method: "POST", Path: "/v1/blog/assets/upload", Handler: uploadHandler(s)},
		{Method: "GET", Path: "/v1/blog/assets", Handler: readAssetHandler(s)},
		{Method: "HEAD", Path: "/v1/blog/assets", Handler: readAssetHandler(s)}}, rest.WithTimeout(10*time.Minute), rest.WithMaxBytes((50<<20)+(1<<20)))
}
func uploadHandler(s *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, (50<<20)+(1<<20))
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			httpx.ErrorCtx(r.Context(), w, status.Error(codes.InvalidArgument, "invalid multipart upload"))
			return
		}
		defer r.MultipartForm.RemoveAll()
		body, header, err := r.FormFile("file")
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, status.Error(codes.InvalidArgument, "file is required"))
			return
		}
		defer body.Close()
		stream, err := s.RPC.UploadAsset(rpcx.Outgoing(r.Context(), s.Config.ServiceKey))
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		err = stream.Send(&pb.StreamChunk{Name: header.Filename, Size: header.Size, Kind: r.FormValue("kind")})
		if err == nil {
			err = streamio.Copy(r.Context(), body, func(data []byte) error { return stream.Send(&pb.StreamChunk{Data: data}) })
		}
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		out, err := stream.CloseAndRecv()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		value, err := protoconv.Value(out)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, map[string]any{"code": 0, "message": "ok", "data": value})
	}
}
