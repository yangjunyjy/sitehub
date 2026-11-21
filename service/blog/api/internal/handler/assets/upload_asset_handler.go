package assets

import (
	"net/http"
	"sitehub/service/blog/api/internal/logic"
	"sitehub/service/blog/api/internal/payload"
	"sitehub/service/blog/api/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func UploadAssetHandler(s *svc.ServiceContext) http.HandlerFunc {
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
