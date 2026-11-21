package settings

import (
	"net/http"

	"sitehub/service/rbac/api/internal/logic/settings"
	"sitehub/service/rbac/api/internal/svc"
	"sitehub/service/rbac/api/internal/transport"
)

func GetEmailSettingsHandler(services *svc.ServiceContext) http.HandlerFunc {
	return transport.WithoutRequest(services, settings.GetEmailSettings)
}
