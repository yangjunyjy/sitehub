package captcha

import (
	"net/http"

	"sitehub/service/rbac/api/internal/logic/captcha"
	"sitehub/service/rbac/api/internal/svc"
	"sitehub/service/rbac/api/internal/transport"
)

func CaptchaHandler(services *svc.ServiceContext) http.HandlerFunc {
	return transport.WithoutRequest(services, captcha.Captcha)
}
