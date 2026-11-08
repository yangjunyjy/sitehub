package svc

import (
	"context"
	coreconfig "sitehub/common/config"
	"sitehub/common/infrastructure/smtp"
	"sitehub/common/pkg/token"
	"sitehub/common/rpcx"
	"sitehub/common/runtimeconf"
	settingsrepo "sitehub/common/settings/repository/impl/mysql"
	settings "sitehub/common/settings/service"
	business "sitehub/service/rbac"
	"sitehub/service/rbac/rpc/internal/config"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config config.Config
	Core   *business.Module
	DB     *gorm.DB
	Cache  *redis.Client
	TOTP   *token.TOTP
}

func NewServiceContext(c config.Config) *ServiceContext {
	if len(c.ServiceKey) < 32 {
		panic("ServiceKey must contain at least 32 characters")
	}
	if len(c.AuthSecret) < 32 {
		panic("AuthSecret must contain at least 32 characters")
	}
	db, err := c.Database.Open()
	if err != nil {
		panic(err)
	}
	if err = runtimeconf.RequireOwner(context.Background(), db, "rbac"); err != nil {
		panic(err)
	}
	s := &ServiceContext{Config: c, DB: db}
	cache := redis.NewClient(&redis.Options{Addr: c.Cache.Addr, Password: c.Cache.Password, DB: c.Cache.DB})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err = cache.Ping(ctx).Err(); err != nil {
		panic(err)
	}
	cfg := &coreconfig.Config{}
	cfg.Auth.JWTSecret = c.AuthSecret
	cfg.Session.MaxAge = c.SessionTTL
	cfg.OnceToken.ExpireTime = c.OnceTokenTTL
	cfg.Auth.TOTP.Enabled = c.TOTP.Enabled
	cfg.Auth.TOTP.Secret = c.TOTP.Secret
	cfg.Auth.TOTP.TimeStep = c.TOTP.TimeStep
	cfg.Auth.TOTP.WindowSize = c.TOTP.WindowSize
	cfg.EmailServer.SMTPHost = c.Mail.Host
	cfg.EmailServer.SMTPPort = c.Mail.Port
	cfg.EmailServer.FromEmail = c.Mail.Username
	cfg.EmailServer.FromName = c.Mail.FromName
	cfg.EmailServer.Password = c.Mail.Password
	cfg.EmailServer.Length = 6
	cfg.EmailServer.Charset = "0123456789"
	cfg.EmailServer.ExpiresIn = 5 * time.Minute
	cfg.Notifications.BlacklistEmail = c.BlacklistEmail
	core := business.New(db, cache, cfg, settings.ScopeAtomic(settingsrepo.NewSystemConfigRepository(db), "system.email"))
	smtp.SetConfigProvider(core.Email.RuntimeConfig)
	var totp *token.TOTP
	if cfg.Auth.TOTP.Enabled {
		totp, err = token.NewTOTP(&cfg.Auth)
		if err != nil {
			panic(err)
		}
	}
	s.Core = core
	s.Cache = cache
	s.TOTP = totp
	return s
}
func (s *ServiceContext) Close() {
	db, err := s.DB.DB()
	if err == nil {
		_ = db.Close()
	}
	_ = s.Cache.Close()
}

func (s *ServiceContext) CheckClient(ctx context.Context) error {
	ip := rpcx.Incoming(ctx, "x-client-ip")
	if ip == "" {
		return nil
	}
	blocked, err := s.Core.Blacklist.CheckIPBlocked(ip)
	if err != nil {
		return rpcx.Error(err)
	}
	if blocked {
		return status.Error(codes.PermissionDenied, "client address blocked")
	}
	return nil
}
