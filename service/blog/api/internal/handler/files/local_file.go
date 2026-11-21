package files

import (
	"github.com/zeromicro/go-zero/rest/httpx"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"sitehub/service/blog/api/internal/logic"
	"sitehub/service/blog/api/internal/svc"
	"time"
)

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
