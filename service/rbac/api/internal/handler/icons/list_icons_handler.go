package icons

import (
	"net/http"

	"sitehub/service/rbac/api/internal/logic/icons"
	"sitehub/service/rbac/api/internal/svc"
	"sitehub/service/rbac/api/internal/transport"
)

func ListIconsHandler(services *svc.ServiceContext) http.HandlerFunc {
	return transport.WithoutRequest(services, icons.ListIcons)
}
