package roles

import (
	"net/http"

	"sitehub/service/rbac/api/internal/logic/roles"
	"sitehub/service/rbac/api/internal/svc"
	"sitehub/service/rbac/api/internal/transport"
)

func DeleteRoleHandler(services *svc.ServiceContext) http.HandlerFunc {
	return transport.WithRequest(services, roles.DeleteRole)
}
