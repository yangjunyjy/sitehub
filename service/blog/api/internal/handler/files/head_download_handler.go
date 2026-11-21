package files

import (
	"net/http"
	"sitehub/service/blog/api/internal/svc"
)

func HeadDownloadHandler(services *svc.ServiceContext) http.HandlerFunc {
	return DownloadFileHandler(services)
}
