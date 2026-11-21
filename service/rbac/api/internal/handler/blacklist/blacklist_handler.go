package blacklist

import (
	"net/http"

	"sitehub/service/rbac/api/internal/logic/blacklist"
	"sitehub/service/rbac/api/internal/svc"
	"sitehub/service/rbac/api/internal/transport"
)

func BlacklistHandler(services *svc.ServiceContext) http.HandlerFunc {
	return transport.WithoutRequest(services, blacklist.Blacklist)
}
