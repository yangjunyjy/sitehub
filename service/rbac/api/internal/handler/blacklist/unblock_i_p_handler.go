package blacklist

import (
	"net/http"

	"sitehub/service/rbac/api/internal/logic/blacklist"
	"sitehub/service/rbac/api/internal/svc"
	"sitehub/service/rbac/api/internal/transport"
)

func UnblockIPHandler(services *svc.ServiceContext) http.HandlerFunc {
	return transport.WithRequest(services, blacklist.UnblockIP)
}
