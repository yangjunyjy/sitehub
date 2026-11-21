package users

import (
	"net/http"

	"sitehub/service/rbac/api/internal/logic/users"
	"sitehub/service/rbac/api/internal/svc"
	"sitehub/service/rbac/api/internal/transport"
)

func ListUsersHandler(services *svc.ServiceContext) http.HandlerFunc {
	return transport.WithRequest(services, users.ListUsers)
}
