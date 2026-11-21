package captcha

import (
	"net/http"

	"sitehub/service/rbac/api/internal/logic/captcha"
	"sitehub/service/rbac/api/internal/svc"
	"sitehub/service/rbac/api/internal/transport"
)

func VerifyCaptchaHandler(services *svc.ServiceContext) http.HandlerFunc {
	return transport.WithRequest(services, captcha.VerifyCaptcha)
}
