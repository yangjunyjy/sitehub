package icons

import (
	"net/http"

	"sitehub/service/rbac/api/internal/logic/icons"
	"sitehub/service/rbac/api/internal/svc"
	"sitehub/service/rbac/api/internal/transport"
)

func UpdateIconHandler(services *svc.ServiceContext) http.HandlerFunc {
	return transport.WithRequest(services, icons.UpdateIcon)
}
