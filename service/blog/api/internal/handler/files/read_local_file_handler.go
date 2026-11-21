package files

import (
	"net/http"
	"sitehub/service/blog/api/internal/svc"
	"strings"
)

func ReadLocalFileHandler(s *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		location := strings.TrimPrefix(r.URL.Path, "/v1/file/local/")
		if !strings.HasPrefix(location, "images/") && !strings.HasPrefix(location, "attachments/") {
			http.NotFound(w, r)
			return
		}
		serveLocalFile(s, w, r, location, "")
	}
}
