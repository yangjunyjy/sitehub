package handler

import (
	"net/http"
	"sitehub/service/blog/api/internal/logic"
	"sitehub/service/blog/api/internal/payload"
	"sitehub/service/blog/api/internal/svc"
	"sitehub/service/blog/business"

	"github.com/zeromicro/go-zero/rest/httpx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func uploadFileHandler(s *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := s.Require(r.Context(), "file:storage:edit", "admin", true); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, (2<<30)+(1<<20))
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
		if header.Size <= 0 || header.Size > 2<<30 {
			httpx.ErrorCtx(r.Context(), w, status.Error(codes.InvalidArgument, "invalid upload size"))
			return
		}
		var value *business.AttachmentInfo
		switch r.FormValue("kind") {
		case "image":
			var location string
			location, err = s.Core.Uploads.UploadImage(header.Filename, header.Size, body)
			value = &business.AttachmentInfo{URL: location, Name: header.Filename, Size: header.Size}
		case "attachment":
			value, err = s.Core.Uploads.UploadAttachment(header.Filename, header.Size, body)
		case "release":
			value, err = s.Core.Uploads.UploadRelease(header.Filename, header.Size, body)
		default:
			err = status.Error(codes.InvalidArgument, "unknown upload kind")
		}
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, logic.Classify(err))
			return
		}
		response, err := logic.ReplyAs[payload.FileInfo](value, nil)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, response)
	}
}
