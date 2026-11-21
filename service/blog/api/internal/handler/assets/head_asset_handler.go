package assets

import (
	"net/http"
	"sitehub/service/blog/api/internal/svc"
)

func HeadAssetHandler(services *svc.ServiceContext) http.HandlerFunc {
	return ReadAssetHandler(services)
}
