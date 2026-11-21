package files

import (
	"net/http"
	"sitehub/service/blog/api/internal/svc"
)

func HeadLocalFileHandler(services *svc.ServiceContext) http.HandlerFunc {
	return ReadLocalFileHandler(services)
}
