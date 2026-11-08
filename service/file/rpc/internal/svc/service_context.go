package svc

import (
	"context"
	coreconfig "sitehub/common/config"
	"sitehub/common/rpcx"
	"sitehub/common/runtimeconf"
	settingsrepo "sitehub/common/settings/repository/impl/mysql"
	settings "sitehub/common/settings/service"
	business "sitehub/service/file"
	"sitehub/service/file/rpc/internal/config"
	rbacpb "sitehub/service/rbac/rpc/pb/rbac"

	"github.com/zeromicro/go-zero/zrpc"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config config.Config
	Core   *business.Module
	DB     *gorm.DB
	Rbac   rbacpb.RbacClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	if len(c.ServiceKey) < 32 {
		panic("ServiceKey must contain at least 32 characters")
	}
	if len(c.StorageSecret) < 32 {
		panic("StorageSecret must contain at least 32 characters")
	}
	db, err := c.Database.Open()
	if err != nil {
		panic(err)
	}
	if err = runtimeconf.RequireOwner(context.Background(), db, "file"); err != nil {
		panic(err)
	}
	s := &ServiceContext{Config: c, DB: db}
	c.RbacRPC.Middlewares.Duration = false
	s.Rbac = rbacpb.NewRbacClient(zrpc.MustNewClient(c.RbacRPC).Conn())
	cfg := &coreconfig.Config{}
	cfg.Auth.JWTSecret = c.StorageSecret
	cfg.Upload.Type = "local"
	cfg.Upload.Local.UploadPath = c.StorageRoot
	cfg.Upload.URLPrefix = c.PublicBaseURL
	s.Core = business.New(cfg, settings.Scope(settings.NewSystemConfigService(settingsrepo.NewSystemConfigRepository(db)), "website.storage"))
	return s
}
func (s *ServiceContext) Close() {
	db, err := s.DB.DB()
	if err == nil {
		_ = db.Close()
	}
}
func (s *ServiceContext) Require(ctx context.Context, permission, audience string, secure bool) (*rbacpb.Principal, error) {
	return s.Rbac.Authorize(rpcx.Outgoing(ctx, s.Config.ServiceKey), &rbacpb.AuthorizeRequest{Token: rpcx.Incoming(ctx, "authorization"), Permission: permission, Audience: audience, Secure: secure, Totp: rpcx.Incoming(ctx, "x-totp-token")})
}

func (s *ServiceContext) CheckClient(ctx context.Context) error {
	if rpcx.Incoming(ctx, "x-client-ip") == "" {
		return nil
	}
	_, err := s.Rbac.CheckClient(rpcx.Outgoing(ctx, s.Config.ServiceKey), &rbacpb.Empty{})
	return err
}
