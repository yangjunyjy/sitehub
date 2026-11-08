package handler

import (
	"context"
	"io"
	"mime"
	"sitehub/common/rpcx"
	"sitehub/common/streamio"
	"sitehub/service/file/api/internal/svc"
	pb "sitehub/service/file/rpc/pb/file"
	productpb "sitehub/service/product/rpc/pb/product"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/rest/httpx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func localHandler(s *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		location := strings.TrimPrefix(r.URL.Path, "/v1/file/local/")
		if !strings.HasPrefix(location, "images/") && !strings.HasPrefix(location, "attachments/") {
			http.NotFound(w, r)
			return
		}
		serveLocal(s, w, r, location, "")
	}
}
func downloadHandler(s *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Product checks the member session and current publication state on every download.
		value, err := s.Product.ResolveDownload(rpcx.Outgoing(r.Context(), s.Config.ServiceKey), &productpb.DownloadRequest{Id: r.URL.Query().Get("id")})
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		address, err := url.Parse(value.Url)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, status.Error(codes.Internal, "invalid published URL"))
			return
		}
		if address.IsAbs() {
			if address.Scheme != "http" && address.Scheme != "https" {
				http.NotFound(w, r)
				return
			}
			http.Redirect(w, r, address.String(), http.StatusTemporaryRedirect)
			return
		}
		location := strings.TrimPrefix(address.Path, "/v1/file/local/")
		if !strings.HasPrefix(location, "releases/") {
			http.NotFound(w, r)
			return
		}
		name := path.Base(value.FileName)
		if name == "." || name == "/" || name == "" {
			name = path.Base(location)
		}
		serveLocal(s, w, r, location, name)
	}
}
func serveLocal(s *svc.ServiceContext, w http.ResponseWriter, r *http.Request, location, downloadName string) {
	open := func(offset, length int64) (*pb.StreamChunk, io.ReadCloser, error) {
		ctx, cancel := context.WithCancel(r.Context())
		stream, err := s.RPC.ReadLocal(rpcx.Outgoing(ctx, s.Config.ServiceKey), &pb.LocationRequest{Location: location, Offset: offset, Length: length})
		if err != nil {
			cancel()
			return nil, nil, err
		}
		meta, err := stream.Recv()
		if err != nil {
			cancel()
			return nil, nil, err
		}
		body := &streamio.Reader{Cancel: cancel, Next: func() ([]byte, error) {
			chunk, err := stream.Recv()
			if err != nil {
				return nil, err
			}
			return chunk.Data, nil
		}}
		return meta, body, nil
	}
	meta, body, err := open(0, 1)
	if err != nil {
		httpx.ErrorCtx(r.Context(), w, err)
		return
	}
	body.Close()
	reader := &streamio.Seeker{Size: meta.Size, Open: func(offset, length int64) (io.ReadCloser, error) {
		_, body, err := open(offset, length)
		return body, err
	}}
	defer reader.Close()
	w.Header().Set("Content-Type", meta.Mime)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if downloadName != "" {
		w.Header().Set("Cache-Control", "private, no-store")
		w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": downloadName}))
	}
	http.ServeContent(w, r, meta.Name, time.Time{}, reader)
}
