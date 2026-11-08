package svc

import (
	"context"
	"sitehub/common/rpcx"
	"sitehub/common/runtimeconf"
	settingsrepo "sitehub/common/settings/repository/impl/mysql"
	settings "sitehub/common/settings/service"
	fileclient "sitehub/service/file/client"
	filepb "sitehub/service/file/rpc/pb/file"
	business "sitehub/service/product"
	"sitehub/service/product/rpc/internal/config"
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
	db, err := c.Database.Open()
	if err != nil {
		panic(err)
	}
	if err = runtimeconf.RequireOwner(context.Background(), db, "product"); err != nil {
		panic(err)
	}
	s := &ServiceContext{Config: c, DB: db}
	c.RbacRPC.Middlewares.Duration = false
	s.Rbac = rbacpb.NewRbacClient(zrpc.MustNewClient(c.RbacRPC).Conn())
	files := fileclient.New(filepb.NewFileClient(zrpc.MustNewClient(c.FileRPC).Conn()), c.ServiceKey)
	s.Core = business.New(db, settings.Scope(settings.NewSystemConfigService(settingsrepo.NewSystemConfigRepository(db)), "website.content", "website.downloads"), files)
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
