package svc

import (
	"context"
	"fmt"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"gorm.io/gorm"

	"sitehub/common/rpcx"
	"sitehub/common/runtimeconf"
	"sitehub/service/product"
	"sitehub/service/product/api/internal/config"
	fileclient "sitehub/service/blog/client"
	rbacpb "sitehub/service/rbac/rpc/pb/rbac"
	settings "sitehub/common/settings/service"
	settingsrepo "sitehub/common/settings/repository/impl/mysql"
)

type ServiceContext struct {
	Config         config.Config
	Core           *product.Module
	DB             *gorm.DB
	Rbac           rbacpb.RbacClient
	rbacConnection *grpc.ClientConn
}

func NewServiceContext(configuration config.Config) (*ServiceContext, error) {
	if len(configuration.ServiceKey) < 32 {
		return nil, fmt.Errorf("ServiceKey must contain at least 32 characters")
	}
	database, err := configuration.Database.Open()
	if err != nil {
		return nil, err
	}
	service := &ServiceContext{Config: configuration, DB: database}
	if err = runtimeconf.RequireOwner(context.Background(), database, "product"); err != nil {
		service.Close()
		return nil, err
	}
	configuration.RbacRPC.Middlewares.Duration = false
	rbac, err := zrpc.NewClient(configuration.RbacRPC)
	if err != nil {
		service.Close()
		return nil, err
	}
	service.rbacConnection = rbac.Conn()
	service.Rbac = rbacpb.NewRbacClient(rbac.Conn())
	store := settings.Scope(settings.NewSystemConfigService(settingsrepo.NewSystemConfigRepository(database)), "website.content")
	service.Core = product.New(database, store, fileclient.New(configuration.BlogURL, configuration.ServiceKey))
	return service, nil
}

func (service *ServiceContext) Close() {
	if service.rbacConnection != nil {
		_ = service.rbacConnection.Close()
	}
	if service.DB != nil {
		if pool, err := service.DB.DB(); err == nil {
			_ = pool.Close()
		}
	}
}

func (service *ServiceContext) Require(ctx context.Context, permission, audience string, secure bool) (*rbacpb.Principal, error) {
	return service.Rbac.Authorize(rpcx.Outgoing(ctx, service.Config.ServiceKey), &rbacpb.AuthorizeRequest{Token: rpcx.HTTPValue(ctx, "authorization"), Totp: rpcx.HTTPValue(ctx, "x-totp-token"), Permission: permission, Audience: audience, Secure: secure})
}

func (service *ServiceContext) CheckClient(ctx context.Context) error {
	if rpcx.HTTPValue(ctx, "x-client-ip") == "" {
		return nil
	}
	_, err := service.Rbac.CheckClient(rpcx.Outgoing(ctx, service.Config.ServiceKey), &rbacpb.Empty{})
	return err
}

func (service *ServiceContext) ClientPolicy(next http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/internal/product/download" && request.URL.Path != "/health" {
			if err := service.CheckClient(request.Context()); err != nil {
				httpx.ErrorCtx(request.Context(), writer, err)
				return
			}
		}
		next(writer, request)
	}
}
