package members

import (
	"net/http"

	"sitehub/service/rbac/api/internal/logic/members"
	"sitehub/service/rbac/api/internal/svc"
	"sitehub/service/rbac/api/internal/transport"
)

func CurrentMemberHandler(services *svc.ServiceContext) http.HandlerFunc {
	return transport.WithoutRequest(services, members.CurrentMember)
}
