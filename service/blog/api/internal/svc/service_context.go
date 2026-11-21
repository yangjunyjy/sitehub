package svc

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/rest/httpx"
	"net/http"
	coreconfig "sitehub/common/config"
	"sitehub/common/rpcx"
	"sitehub/common/runtimeconf"
	settingsrepo "sitehub/common/settings/repository/impl/mysql"
	settings "sitehub/common/settings/service"
	business "sitehub/service/blog"
	"sitehub/service/blog/api/internal/config"
	productclient "sitehub/service/product/client"
	productcontract "sitehub/service/product/contract"
	rbacpb "sitehub/service/rbac/rpc/pb/rbac"
	"strings"
	"sync"

	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type DownloadResolver interface {
	ResolveDownload(context.Context, string) (productcontract.ProductDownload, error)
}

type ServiceContext struct {
	Product   DownloadResolver
	Config    config.Config
	Core      *business.Module
	DB        *gorm.DB
	Rbac      rbacpb.RbacClient
	StorageDB *gorm.DB
	rbacConn  *grpc.ClientConn
	close     sync.Once
}

// New builds one Blog module and owns all of its infrastructure connections.
func New(c config.Config) (*ServiceContext, error) {
	if len(c.ServiceKey) < 32 {
		return nil, fmt.Errorf("ServiceKey must contain at least 32 characters")
	}
	if len(c.StorageSecret) < 32 {
		return nil, fmt.Errorf("StorageSecret must contain at least 32 characters")
	}
	s := &ServiceContext{Config: c, Product: productclient.New(c.ProductURL, c.ServiceKey)}
	ready := false
	defer func() {
		if !ready {
			s.Close()
		}
	}()
	db, err := c.Database.Open()
	if err != nil {
		return nil, fmt.Errorf("open blog database: %w", err)
	}
	s.DB = db
	if err = runtimeconf.RequireOwner(context.Background(), db, "blog"); err != nil {
		return nil, err
	}
	s.StorageDB, err = c.FileDatabase.Open()
	if err != nil {
		return nil, fmt.Errorf("open storage database: %w", err)
	}
	if err = runtimeconf.RequireOwner(context.Background(), s.StorageDB, "file"); err != nil {
		return nil, err
	}
	c.RbacRPC.Middlewares.Duration = false
	rbac, err := zrpc.NewClient(c.RbacRPC)
	if err != nil {
		return nil, fmt.Errorf("connect to RBAC: %w", err)
	}
	s.rbacConn = rbac.Conn()
	s.Rbac = rbacpb.NewRbacClient(s.rbacConn)
	uploadConfig := coreconfig.UploadConfig{Type: "local", URLPrefix: c.PublicBaseURL}
	uploadConfig.Local.UploadPath = c.StorageRoot
	store := settings.Scope(settings.NewSystemConfigService(settingsrepo.NewSystemConfigRepository(s.StorageDB)), "website.storage")
	s.Core = business.New(db, uploadConfig, store, c.StorageSecret)
	ready = true
	return s, nil
}

func (s *ServiceContext) Health(ctx context.Context) error {
	if s.DB == nil || s.StorageDB == nil {
		return fmt.Errorf("blog and file databases must be initialized")
	}
	for _, entry := range []struct {
		name string
		db   *gorm.DB
	}{{"blog", s.DB}, {"file", s.StorageDB}} {
		pool, err := entry.db.DB()
		if err != nil {
			return fmt.Errorf("%s database unavailable: %w", entry.name, err)
		}
		if err = pool.PingContext(ctx); err != nil {
			return fmt.Errorf("%s database unavailable: %w", entry.name, err)
		}
	}
	return nil
}
func (s *ServiceContext) Close() {
	s.close.Do(func() {
		if s.StorageDB != nil {
			if db, err := s.StorageDB.DB(); err == nil {
				_ = db.Close()
			}
		}
		if s.DB != nil {
			if db, err := s.DB.DB(); err == nil {
				_ = db.Close()
			}
		}
		if s.rbacConn != nil {
			_ = s.rbacConn.Close()
		}
	})
}
func (s *ServiceContext) Require(ctx context.Context, permission, audience string, secure bool) (*rbacpb.Principal, error) {
	outgoing := rpcx.Outgoing(ctx, s.Config.ServiceKey)
	request := &rbacpb.AuthorizeRequest{Token: rpcx.HTTPValue(ctx, "authorization"), Permission: permission, Audience: audience, Secure: secure, Totp: rpcx.HTTPValue(ctx, "x-totp-token")}
	principal, err := s.Rbac.Authorize(outgoing, request)
	if status.Code(err) != codes.PermissionDenied || status.Convert(err).Message() != "permission denied" {
		return principal, err
	}
	legacy := legacyPermission(permission)
	if legacy == "" {
		return nil, err
	}
	request.Permission = legacy
	return s.Rbac.Authorize(outgoing, request)
}

// Member resolves the site session before any blog transaction starts.
func (s *ServiceContext) Member(ctx context.Context) (*rbacpb.Principal, error) {
	if rpcx.HTTPValue(ctx, "authorization") == "" {
		return nil, status.Error(codes.Unauthenticated, "请先登录")
	}
	principal, err := s.Require(ctx, "", "member", false)
	if err != nil {
		return nil, err
	}
	if !principal.GetMember() || principal.GetId() == 0 {
		return nil, status.Error(codes.Unauthenticated, "请重新登录")
	}
	return principal, nil
}

func (s *ServiceContext) CheckClient(ctx context.Context) error {
	if rpcx.HTTPValue(ctx, "x-client-ip") == "" {
		return nil
	}
	_, err := s.Rbac.CheckClient(rpcx.Outgoing(ctx, s.Config.ServiceKey), &rbacpb.Empty{})
	return err
}

func (s *ServiceContext) ClientPolicy(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/readyz" && r.URL.Path != "/healthz" && r.URL.Path != "/internal/governance" && r.URL.Path != "/health" && !strings.HasPrefix(r.URL.Path, "/internal/blog/") {
			if err := s.CheckClient(r.Context()); err != nil {
				httpx.ErrorCtx(r.Context(), w, err)
				return
			}
		}
		next(w, r)
	}
}

func legacyPermission(permission string) string {
	switch permission {
	case "blog:category:create":
		return "blog:category:add"
	case "blog:category:update":
		return "blog:category:edit"
	case "blog:tag:create":
		return "blog:tag:add"
	case "blog:tag:update":
		return "blog:tag:edit"
	case "file:storage:list", "file:storage:edit":
		return "website:storage:edit"
	case "blog:statistics:list":
		return "blog:article:list"
	default:
		return ""
	}
}
