package console

import (
	"context"
	"io"
	"mime"
	"net/http"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"sitehub/common/rpcx"
	pb "sitehub/service/rbac/rpc/pb/rbac"
)

const sessionCookie = "sitehub_admin_session"

// Use only the cookie for the browser identity. RPC also receives the peer IP and agent
// so login throttling and session metadata use the same inputs as the admin API.
func (c *console) authRoutes(mux *http.ServeMux) {
	withPeer := rpcx.HTTPContextWithProxies(c.trustedProxies)
	for path, handler := range map[string]http.HandlerFunc{
		"GET /api/auth/captcha": c.captcha, "POST /api/auth/captcha/verify": c.verifyCaptcha,
		"POST /api/auth/code": c.sendCode, "POST /api/auth/login": c.login, "POST /api/auth/logout": c.logout,
	} {
		authenticatedRequest := withPeer(handler)
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "POST" && r.Header.Get("Origin") != c.origin {
				problem(w, 403, "请求来源与控制台配置不一致")
				return
			}
			r.Header.Del("Authorization")
			if cookies := r.CookiesNamed(sessionCookie); len(cookies) == 1 {
				r.Header.Set("Authorization", "Bearer "+cookies[0].Value)
			}
			authenticatedRequest(w, r)
		})
	}
}
func (c *console) authContext(r *http.Request) (context.Context, context.CancelFunc) {
	return context.WithTimeout(rpcx.Outgoing(r.Context(), c.serviceKey), 20*time.Second)
}
func readAuth(w http.ResponseWriter, r *http.Request, in proto.Message) bool {
	typ, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || typ != "application/json" {
		problem(w, 415, "请提交 JSON")
		return false
	}
	data, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 16<<10))
	if err != nil || protojson.Unmarshal(data, in) != nil {
		problem(w, 400, "登录请求格式有误")
		return false
	}
	return true
}
func authError(w http.ResponseWriter, err error, invalid string) {
	switch status.Code(err) {
	case codes.InvalidArgument, codes.FailedPrecondition:
		problem(w, 400, invalid)
	case codes.Unauthenticated:
		problem(w, 401, invalid)
	case codes.PermissionDenied:
		problem(w, 403, "当前账号无权访问网关控制台")
	case codes.ResourceExhausted:
		problem(w, 429, "请求过于频繁，请稍后重试")
	default:
		problem(w, 503, "认证服务暂时不可用，请稍后重试")
	}
}
func (c *console) captcha(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := c.authContext(r)
	defer cancel()
	out, err := c.auth.Captcha(ctx, &pb.Empty{})
	if err != nil {
		authError(w, err, "获取拼图失败")
		return
	}
	reply(w, 200, map[string]any{"captchaKey": out.CaptchaKey, "backgroundImage": out.BackgroundImage, "sliderImage": out.SliderImage, "y": out.Y, "imgWidth": out.ImgWidth, "imgHeight": out.ImgHeight, "pieceSize": out.PieceSize, "piecePadding": out.PiecePadding})
}
func (c *console) verifyCaptcha(w http.ResponseWriter, r *http.Request) {
	var in pb.CaptchaVerifyRequest
	if !readAuth(w, r, &in) {
		return
	}
	ctx, cancel := c.authContext(r)
	defer cancel()
	out, err := c.auth.VerifyCaptcha(ctx, &in)
	if err != nil {
		authError(w, err, "拼图已失效，请重新验证")
		return
	}
	reply(w, 200, map[string]bool{"value": out.Value})
}
func (c *console) sendCode(w http.ResponseWriter, r *http.Request) {
	var in pb.SendCodeRequest
	if !readAuth(w, r, &in) {
		return
	}
	ctx, cancel := c.authContext(r)
	defer cancel()
	if _, err := c.auth.SendAdminCode(ctx, &in); err != nil {
		authError(w, err, "邮件未能发送，请检查邮箱并重新验证拼图")
		return
	}
	w.WriteHeader(204)
}
func (c *console) setSession(w http.ResponseWriter, token string, age int) {
	http.SetCookie(w, &http.Cookie{Name: sessionCookie, Value: token, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: strings.HasPrefix(c.origin, "https://"), MaxAge: age})
}
func (c *console) login(w http.ResponseWriter, r *http.Request) {
	var in pb.SysLoginRequest
	if !readAuth(w, r, &in) {
		return
	}
	ctx, cancel := c.authContext(r)
	defer cancel()
	session, err := c.auth.Login(ctx, &in)
	if err != nil {
		authError(w, err, "账号、密码或邮箱验证码有误")
		return
	}
	age := int(min(int64(86400), session.ExpiresAt-time.Now().Unix()))
	if session.Token == "" || age <= 0 {
		problem(w, 502, "未取得有效会话")
		return
	}
	c.setSession(w, session.Token, age)
	w.WriteHeader(204)
}
func (c *console) logout(w http.ResponseWriter, r *http.Request) {
	if rpcx.HTTPValue(r.Context(), "authorization") != "" {
		ctx, cancel := c.authContext(r)
		defer cancel()
		if _, err := c.auth.Logout(ctx, &pb.Empty{}); err != nil && status.Code(err) != codes.Unauthenticated {
			authError(w, err, "退出未完成，请重试")
			return
		}
	}
	c.setSession(w, "", -1)
	w.WriteHeader(204)
}
