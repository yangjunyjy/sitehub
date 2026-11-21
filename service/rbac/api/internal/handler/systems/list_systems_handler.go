package systems

import (
	"net/http"

	"sitehub/service/rbac/api/internal/logic/systems"
	"sitehub/service/rbac/api/internal/svc"
	"sitehub/service/rbac/api/internal/transport"
)

func ListSystemsHandler(services *svc.ServiceContext) http.HandlerFunc {
	return transport.WithoutRequest(services, systems.ListSystems)
}
