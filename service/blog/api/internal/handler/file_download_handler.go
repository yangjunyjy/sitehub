package handler

import (
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sitehub/service/blog/api/internal/logic"
	"sitehub/service/blog/api/internal/svc"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/rest/httpx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func readLocalFileHandler(s *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		location := strings.TrimPrefix(r.URL.Path, "/v1/file/local/")
		if !strings.HasPrefix(location, "images/") && !strings.HasPrefix(location, "attachments/") {
			http.NotFound(w, r)
			return
		}
		serveLocalFile(s, w, r, location, "")
	}
}
func downloadFileHandler(s *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Product checks the member session and current publication state on every download.
		value, err := s.Product.ResolveDownload(r.Context(), r.URL.Query().Get("id"))
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		address, err := url.Parse(value.URL)
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
		serveLocalFile(s, w, r, location, name)
	}
}
func serveLocalFile(s *svc.ServiceContext, w http.ResponseWriter, r *http.Request, location, downloadName string) {
	name, err := s.Core.Uploads.ServeLocalPath(location)
	if err != nil {
		httpx.ErrorCtx(r.Context(), w, logic.Classify(err))
		return
	}
	file, err := os.Open(name)
	if err != nil {
		httpx.ErrorCtx(r.Context(), w, logic.Classify(err))
		return
	}
	defer file.Close()
	w.Header().Set("Content-Type", mime.TypeByExtension(filepath.Ext(name)))
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if downloadName != "" {
		w.Header().Set("Cache-Control", "private, no-store")
		w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": downloadName}))
	}
	http.ServeContent(w, r, filepath.Base(name), time.Time{}, file)
}
