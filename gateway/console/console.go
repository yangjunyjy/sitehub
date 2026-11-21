package console

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sitehub/gateway/plugin"
	"slices"
	"strconv"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"sitehub/common/rpcx"
	"sitehub/gateway/publisher"
	pb "sitehub/service/rbac/rpc/pb/rbac"
	ui "sitehub/web/gateway-admin"
)

const (
	readConfig     = "gateway:config:read"
	publishConfig  = "gateway:config:publish"
	rollbackConfig = "gateway:config:rollback"
)

type Config struct {
	TrustedProxies []string `json:",optional"`
	Listen         string   `json:",optional"`
	Origin         string   `json:",optional"`
	RBACAddress    string   `json:",optional"`
	ServiceKey     string   `json:",optional"`
}
type console struct {
	trustedProxies []string
	auth           pb.RbacClient
	serviceKey     string
	plugins        *plugin.Registry
	client         *publisher.Client
	dir, origin    string
	authorize      func(context.Context, *pb.AuthorizeRequest) (*pb.Principal, error)
}

// New builds the browser and management API handler for one gateway.
func New(cfg Config, client *publisher.Client, plugins *plugin.Registry, dir string) (http.Handler, func() error, error) {
	if err := cfg.Validate(); err != nil {
		return nil, nil, err
	}
	conn, err := grpc.NewClient(cfg.RBACAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	auth := pb.NewRbacClient(conn)
	c := console{trustedProxies: cfg.TrustedProxies, client: client, dir: dir, origin: cfg.Origin, auth: auth, serviceKey: cfg.ServiceKey, plugins: plugins,
		authorize: func(ctx context.Context, in *pb.AuthorizeRequest) (*pb.Principal, error) {
			return auth.Authorize(rpcx.Outgoing(ctx, cfg.ServiceKey), in)
		}}
	return c.handler(), conn.Close, nil
}
func (cfg Config) Validate() error {
	for _, cidr := range cfg.TrustedProxies {
		if _, _, err := net.ParseCIDR(cidr); err != nil {
			return fmt.Errorf("invalid Console.TrustedProxies CIDR")
		}
	}
	origin, err := url.Parse(cfg.Origin)
	if err != nil || origin.Hostname() == "" || (origin.Scheme != "https" && origin.Scheme != "http") || cfg.Origin != origin.Scheme+"://"+origin.Host || len(cfg.ServiceKey) < 32 || cfg.RBACAddress == "" || cfg.Listen == "" {
		return fmt.Errorf("console requires Listen, an exact HTTP(S) Origin, RBACAddress and a ServiceKey of at least 32 characters")
	}
	return nil
}
func reply(w http.ResponseWriter, code int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(data)
}
func problem(w http.ResponseWriter, code int, message string) {
	reply(w, code, map[string]string{"message": message})
}
func (c *console) handler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /", ui.Handler())
	c.authRoutes(mux)
	mux.HandleFunc("GET /api/workspace", c.guard(readConfig, false, c.workspace))
	mux.HandleFunc("POST /api/preview", c.guard(readConfig, false, c.change))
	mux.HandleFunc("POST /api/publish", c.guard(publishConfig, true, c.change))
	mux.HandleFunc("POST /api/rollback", c.guard(rollbackConfig, true, c.change))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'; object-src 'none'; base-uri 'none'; frame-ancestors 'none'; form-action 'self'")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Cache-Control", "no-store")
		mux.ServeHTTP(w, r)
	})
}

type consoleHandler func(http.ResponseWriter, *http.Request, *pb.Principal)

func (c *console) guard(permission string, secure bool, next consoleHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && r.Header.Get("Origin") != c.origin {
			problem(w, 403, "请求来源与控制台配置不一致")
			return
		}
		cookies := r.CookiesNamed(sessionCookie)
		if len(cookies) != 1 || cookies[0].Value == "" {
			problem(w, 401, "请登录网关控制台")
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		actor, err := c.authorize(ctx, &pb.AuthorizeRequest{Audience: "admin", Token: cookies[0].Value, Permission: permission, Secure: secure, Totp: r.Header.Get("X-TOTP-Token")})
		if err != nil {
			switch status.Code(err) {
			case codes.Unauthenticated:
				problem(w, 401, "后台会话已失效，请重新登录")
			case codes.PermissionDenied:
				if status.Convert(err).Message() == "invalid TOTP" {
					problem(w, 403, "请输入有效的动态验证码")
				} else {
					problem(w, 403, "当前账号没有此操作的权限")
				}
			default:
				problem(w, 503, "暂时无法校验后台权限")
			}
			return
		}
		next(w, r, actor)
	}
}
func (c *console) workspace(w http.ResponseWriter, r *http.Request, actor *pb.Principal) {
	state, err := c.client.Snapshot(r.Context())
	if err != nil {
		problem(w, 502, "无法读取网关配置，请检查控制地址、密钥和管理网段")
		return
	}
	history, err := releaseHistory(c.dir)
	if err != nil {
		problem(w, 500, "无法读取发布记录，请检查 StateDir")
		return
	}
	can := func(p string) bool {
		return !state.FileManaged && (actor.SuperAdmin || slices.Contains(actor.Permissions, "*") || slices.Contains(actor.Permissions, p))
	}
	reply(w, 200, map[string]any{"state": state, "history": history, "actor": actor.Name, "publish": can(publishConfig), "rollback": can(rollbackConfig)})
}
func actorName(p *pb.Principal) string { return p.Name + " (#" + strconv.FormatUint(p.Id, 10) + ")" }
