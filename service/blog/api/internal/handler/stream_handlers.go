package handler

import (
	"net/http"
	"sitehub/service/blog/api/internal/logic"
	"sitehub/service/blog/api/internal/payload"
	"sitehub/service/blog/api/internal/svc"
	"time"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func RegisterStreamHandlers(server *rest.Server, s *svc.ServiceContext) {
	server.AddRoutes([]rest.Route{{Method: "POST", Path: "/v1/blog/assets/upload", Handler: uploadAssetHandler(s)},
		{Method: "GET", Path: "/v1/blog/assets", Handler: readAssetHandler(s)},
		{Method: "HEAD", Path: "/v1/blog/assets", Handler: readAssetHandler(s)}}, rest.WithTimeout(10*time.Minute), rest.WithMaxBytes((50<<20)+(1<<20)))
	server.AddRoutes([]rest.Route{{Method: "POST", Path: "/v1/file/upload", Handler: uploadFileHandler(s)},
		{Method: "GET", Path: "/v1/file/download", Handler: downloadFileHandler(s)},
		{Method: "HEAD", Path: "/v1/file/download", Handler: downloadFileHandler(s)},
		{Method: "GET", Path: "/v1/file/local/:kind/:name", Handler: readLocalFileHandler(s)},
		{Method: "HEAD", Path: "/v1/file/local/:kind/:name", Handler: readLocalFileHandler(s)}}, rest.WithTimeout(10*time.Minute), rest.WithMaxBytes((2<<30)+(1<<20)))
}
func uploadAssetHandler(s *svc.ServiceContext) http.HandlerFunc {
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
		principal, err := s.Require(r.Context(), "", "admin", true)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		actor := logic.ArticleActor(principal)
		if !actor.CanCreate && !actor.CanUpdate {
			httpx.ErrorCtx(r.Context(), w, status.Error(codes.PermissionDenied, "permission denied"))
			return
		}
		if header.Size <= 0 || header.Size > 50<<20 {
			httpx.ErrorCtx(r.Context(), w, status.Error(codes.InvalidArgument, "invalid upload size"))
			return
		}
		value, err := s.Core.Assets.UploadArticleAsset(r.Context(), actor, header.Filename, header.Size, body)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, logic.Classify(err))
			return
		}
		response, err := logic.ReplyAs[payload.AssetInfo](value, nil)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, response)
	}
}
