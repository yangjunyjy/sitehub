package health

import (
	"net/http"

	"sitehub/service/rbac/api/internal/logic/health"
	"sitehub/service/rbac/api/internal/svc"
	"sitehub/service/rbac/api/internal/transport"
)

func HealthHandler(services *svc.ServiceContext) http.HandlerFunc {
	return transport.WithoutRequest(services, health.Health)
}
