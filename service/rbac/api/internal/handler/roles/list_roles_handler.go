package roles

import (
	"net/http"

	"sitehub/service/rbac/api/internal/logic/roles"
	"sitehub/service/rbac/api/internal/svc"
	"sitehub/service/rbac/api/internal/transport"
)

func ListRolesHandler(services *svc.ServiceContext) http.HandlerFunc {
	return transport.WithRequest(services, roles.ListRoles)
}
