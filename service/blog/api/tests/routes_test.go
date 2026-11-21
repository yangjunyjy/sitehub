package api_test

import (
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/router"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sitehub/common/apiresponse"
	"sitehub/common/rpcx"
	blog "sitehub/service/blog"
	"sitehub/service/blog/api/internal/handler"
	"sitehub/service/blog/api/internal/svc"
	"sitehub/service/blog/business"
	blogmysql "sitehub/service/blog/model/mysql"
	product "sitehub/service/product/contract"
	rbacpb "sitehub/service/rbac/rpc/pb/rbac"
	"strings"
	"sync"
	"testing"
)

type releaseStore struct {
	business.UploadService
	path string
}

func (store releaseStore) ServeLocalPath(location string) (string, error) {
	if location != "releases/app.tgz" {
		return "", os.ErrNotExist
	}
	return store.path, nil
}

type productAPIServer struct {
	mu    sync.Mutex
	calls int
}

func (s *productAPIServer) ResolveDownload(ctx context.Context, id string) (product.ProductDownload, error) {
	s.mu.Lock()
	s.calls++
	s.mu.Unlock()
	if rpcx.HTTPValue(ctx, "authorization") != "member-token" {
		return product.ProductDownload{}, status.Error(codes.Unauthenticated, "member login required")
	}
	if id != "release" {
		return product.ProductDownload{}, status.Error(codes.NotFound, "missing release")
	}
	return product.ProductDownload{URL: "/v1/file/local/releases/app.tgz", FileName: "app.tgz"}, nil
}
func unifiedAPI(t *testing.T, authorizers ...rbacpb.RbacClient) (http.Handler, sqlmock.Sqlmock, *productAPIServer) {
	t.Helper()
	apiresponse.Install()
	connection, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { connection.Close() })
	db, err := gorm.Open(mysql.New(mysql.Config{Conn: connection, SkipInitializeWithVersion: true}), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	release := filepath.Join(t.TempDir(), "app.tgz")
	if err := os.WriteFile(release, []byte("0123456789"), 0600); err != nil {
		t.Fatal(err)
	}
	product := &productAPIServer{}
	ctx := &svc.ServiceContext{DB: db, StorageDB: db, Product: product, Core: &blog.Module{Community: blogmysql.NewCommunityRepository(db), Uploads: releaseStore{path: release}}}
	if len(authorizers) > 0 {
		ctx.Rbac = authorizers[0]
	}
	server, err := rest.NewServer(rest.RestConf{})
	if err != nil {
		t.Fatal(err)
	}
	handler.RegisterHandlers(server, ctx)
	mux := router.NewRouter()
	seen := map[string]bool{}
	for _, route := range server.Routes() {
		key := route.Method + " " + route.Path
		if seen[key] {
			t.Fatalf("duplicate %s", key)
		}
		seen[key] = true
		if err := mux.Handle(route.Method, route.Path, rpcx.HTTPContext(route.Handler)); err != nil {
			t.Fatal(err)
		}
	}
	for _, key := range []string{
		"GET /health", "GET /v1/blog/articles",
		"GET /v1/file/settings", "POST /v1/file/settings",
		"POST /v1/blog/assets/upload", "GET /v1/blog/assets", "HEAD /v1/blog/assets",
		"POST /v1/file/upload", "GET /v1/file/download", "HEAD /v1/file/download",
		"GET /v1/file/local/:kind/:name", "HEAD /v1/file/local/:kind/:name",
		"GET /internal/blog/storage", "PUT /internal/blog/storage",
		"POST /internal/blog/storage/defaults", "POST /internal/blog/storage/validate",
	} {
		if !seen[key] {
			t.Fatalf("missing unified route %s", key)
		}
	}
	return mux, mock, product
}
func TestHTTPHealthIncludesLocalStorage(t *testing.T) {
	mux, _, _ := unifiedAPI(t)
	resp := httptest.NewRecorder()
	mux.ServeHTTP(resp, httptest.NewRequest("GET", "/health", nil))
	if resp.Code != 200 || !strings.Contains(resp.Body.String(), "blog-api") {
		t.Fatal(resp.Body.String())
	}
}
func TestUnifiedAPIFileDownloadsRequireMemberAndKeepRangesAndHead(t *testing.T) {
	handler, _, productService := unifiedAPI(t)
	request := func(method, token, byteRange string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, "/v1/file/download?id=release", nil)
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		if byteRange != "" {
			req.Header.Set("Range", byteRange)
		}
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		return response
	}
	if response := request(http.MethodGet, "", ""); response.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous download: %d %s", response.Code, response.Body.String())
	}
	response := request(http.MethodGet, "member-token", "bytes=2-5")
	if response.Code != http.StatusPartialContent || response.Body.String() != "2345" || response.Header().Get("Content-Range") != "bytes 2-5/10" {
		t.Fatalf("range: %d %s %v", response.Code, response.Body.String(), response.Header())
	}
	if response.Header().Get("Cache-Control") != "private, no-store" || !strings.Contains(response.Header().Get("Content-Disposition"), "app.tgz") {
		t.Fatal(response.Header())
	}
	response = request(http.MethodHead, "member-token", "")
	if response.Code != 200 || response.Body.Len() != 0 || response.Header().Get("Content-Length") != "10" {
		t.Fatalf("HEAD: %d %v", response.Code, response.Header())
	}
	productService.mu.Lock()
	defer productService.mu.Unlock()
	if productService.calls != 3 {
		t.Fatalf("each download must recheck authorization; got %d calls", productService.calls)
	}
}
