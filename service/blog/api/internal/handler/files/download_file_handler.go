package files

import (
	"github.com/zeromicro/go-zero/rest/httpx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
	"net/url"
	"path"
	"sitehub/service/blog/api/internal/svc"
	"strings"
)

func DownloadFileHandler(s *svc.ServiceContext) http.HandlerFunc {
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
