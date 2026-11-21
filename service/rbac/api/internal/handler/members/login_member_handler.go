package members

import (
	"net/http"

	"sitehub/service/rbac/api/internal/logic/members"
	"sitehub/service/rbac/api/internal/svc"
	"sitehub/service/rbac/api/internal/transport"
)

func LoginMemberHandler(services *svc.ServiceContext) http.HandlerFunc {
	return transport.WithRequest(services, members.LoginMember)
}
