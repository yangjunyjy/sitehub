package settings

import (
	"net/http"

	"sitehub/service/rbac/api/internal/logic/settings"
	"sitehub/service/rbac/api/internal/svc"
	"sitehub/service/rbac/api/internal/transport"
)

func SaveEmailSettingsHandler(services *svc.ServiceContext) http.HandlerFunc {
	return transport.WithRequest(services, settings.SaveEmailSettings)
}
