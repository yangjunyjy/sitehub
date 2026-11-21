package menus

import (
	"net/http"

	"sitehub/service/rbac/api/internal/logic/menus"
	"sitehub/service/rbac/api/internal/svc"
	"sitehub/service/rbac/api/internal/transport"
)

func DeleteMenuHandler(services *svc.ServiceContext) http.HandlerFunc {
	return transport.WithRequest(services, menus.DeleteMenu)
}
