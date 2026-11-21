package auth

import (
	"net/http"

	"sitehub/service/rbac/api/internal/logic/auth"
	"sitehub/service/rbac/api/internal/svc"
	"sitehub/service/rbac/api/internal/transport"
)

func CurrentPermissionsHandler(services *svc.ServiceContext) http.HandlerFunc {
	return transport.WithoutRequest(services, auth.CurrentPermissions)
}
