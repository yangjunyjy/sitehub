package console

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"

	"sitehub/common/rpcx"
	"sitehub/gateway/plugin"
	"sitehub/gateway/publisher"
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
	CookiePath     string   `json:",optional"`
}
type console struct {
	Config
	identity *identityStore
	plugins  *plugin.Registry
	client   *publisher.Client
	dir      string
}

func New(cfg Config, client *publisher.Client, plugins *plugin.Registry, dir string) (http.Handler, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	s, err := newIdentityStore(dir)
	if err != nil {
		return nil, err
	}
	c := &console{Config: cfg, identity: s, plugins: plugins, client: client, dir: dir}
	return c.handler(), nil
}

func (cfg Config) Validate() error {
	for _, cidr := range cfg.TrustedProxies {
		if _, _, err := net.ParseCIDR(cidr); err != nil {
			return fmt.Errorf("invalid Console.TrustedProxies CIDR")
		}
	}
	origin, err := url.Parse(cfg.Origin)
	if err != nil || origin.Hostname() == "" || (origin.Scheme != "https" && origin.Scheme != "http") || cfg.Origin != origin.Scheme+"://"+origin.Host || cfg.Listen == "" {
		return fmt.Errorf("console requires Listen and an exact HTTP(S) Origin")
	}
	if path := cfg.CookiePath; path != "" {
		cookie := &http.Cookie{Name: sessionCookie, Path: path}
		if !strings.HasPrefix(path, "/") || cookie.Valid() != nil {
			return fmt.Errorf("Console.CookiePath must be an absolute cookie path")
		}
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
	mux.HandleFunc("GET /api/", http.NotFound)
	mux.HandleFunc("POST /api/auth/login", c.login)
	mux.HandleFunc("POST /api/auth/logout", c.logout)
	mux.HandleFunc("GET /api/history", c.guard(readConfig, c.history))
	mux.HandleFunc("GET /api/workspace", c.guard(readConfig, c.workspace))
	mux.HandleFunc("POST /api/match", c.guard(readConfig, c.match))
	mux.HandleFunc("POST /api/preview", c.guard(readConfig, c.change))
	mux.HandleFunc("POST /api/publish", c.guard(publishConfig, c.change))
	mux.HandleFunc("POST /api/rollback", c.guard(rollbackConfig, c.change))
	withPeer := rpcx.HTTPContextWithProxies(c.TrustedProxies)(mux.ServeHTTP)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'; object-src 'none'; base-uri 'none'; frame-ancestors 'none'; form-action 'self'")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Cache-Control", "no-store")
		if r.Method == http.MethodPost && (len(r.Header.Values("Origin")) != 1 || r.Header.Get("Origin") != c.Origin) {
			problem(w, http.StatusForbidden, "请求来源与控制台配置不一致")
			return
		}
		withPeer(w, r)
	})
}

type consoleHandler func(http.ResponseWriter, *http.Request, account)

func (c *console) guard(permission string, next consoleHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		a, err := c.identity.principal(r)
		if errors.Is(err, os.ErrNotExist) {
			problem(w, http.StatusUnauthorized, "请登录网关控制台")
			return
		}
		if err != nil {
			problem(w, http.StatusServiceUnavailable, "无法读取控制台账号")
			return
		}
		if !a.allows(permission) {
			problem(w, http.StatusForbidden, "当前网关账号没有此操作的权限")
			return
		}
		next(w, r, a)
	}
}

func (c *console) workspace(w http.ResponseWriter, r *http.Request, actor account) {
	state, err := c.client.Workspace(r.Context())
	if err != nil {
		problem(w, 502, "无法读取网关配置，请检查控制地址、密钥和管理网段")
		return
	}
	history := []publisher.ReleaseSummary{}
	if r.URL.Query().Get("history") != "0" {
		history, _, err = c.historyPage(r.Context(), 1, 20, &state)
		if err != nil {
			problem(w, 500, "无法读取发布记录，请检查 StateDir")
			return
		}
	}
	canEdit := !state.FileManaged && !state.DesiredUnavailable
	reply(w, 200, map[string]any{"state": state, "history": history, "actor": actor.Name, "role": actor.Role,
		"publish": canEdit && actor.allows(publishConfig), "rollback": canEdit && actor.allows(rollbackConfig)})
}
