package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/router"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"sitehub/common/rpcx"
	"sitehub/service/product"
	"sitehub/service/product/api/internal/config"
	"sitehub/service/product/api/internal/handler"
	"sitehub/service/product/api/internal/svc"
	business "sitehub/service/product/business"
	"sitehub/service/product/contract"
	rbac "sitehub/service/rbac/rpc/pb/rbac"
)

type downloadAuth struct {
	rbac.RbacClient
	calls int
}

func (auth *downloadAuth) CheckClient(context.Context, *rbac.Empty, ...grpc.CallOption) (*rbac.Empty, error) {
	return &rbac.Empty{}, nil
}

func (auth *downloadAuth) Authorize(ctx context.Context, input *rbac.AuthorizeRequest, options ...grpc.CallOption) (*rbac.Principal, error) {
	auth.calls++
	if input.Token != "member-token" || input.Audience != "member" {
		return nil, status.Error(codes.Unauthenticated, "member login required")
	}
	return &rbac.Principal{Id: 8}, nil
}

type publishedManifest struct {
	business.WebsiteContentService
	reads int
}

func (manifest *publishedManifest) GetPublishedDownload(id string) (contract.ProductDownload, error) {
	manifest.reads++
	if id != "release" {
		return contract.ProductDownload{}, status.Error(codes.NotFound, "withdrawn")
	}
	return contract.ProductDownload{URL: "/v1/file/local/releases/app.zip", FileName: "app.zip"}, nil
}

func TestInternalDownloadRequiresBothServiceAndMemberBeforeResolvingRelease(t *testing.T) {
	const key = "0123456789abcdef0123456789abcdef"
	auth, manifest := &downloadAuth{}, &publishedManifest{}
	service := &svc.ServiceContext{Config: config.Config{ServiceKey: key}, Core: &product.Module{Content: manifest}, Rbac: auth}
	server, err := rest.NewServer(rest.RestConf{})
	if err != nil {
		t.Fatal(err)
	}
	handler.RegisterHandlers(server, service)
	mux := router.NewRouter()
	for _, route := range server.Routes() {
		if err := mux.Handle(route.Method, route.Path, rpcx.HTTPContext(route.Handler)); err != nil {
			t.Fatal(err)
		}
	}
	for _, sample := range []struct {
		key, token, id string
		want           int
	}{{"", "member-token", "release", 401}, {key, "", "release", 401}, {key, "member-token", "withdrawn", 404}, {key, "member-token", "release", 200}} {
		req := httptest.NewRequest(http.MethodGet, "/internal/product/download?id="+sample.id, nil)
		req.Header.Set("X-Service-Key", sample.key)
		req.Header.Set("Authorization", "Bearer "+sample.token)
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, req)
		if response.Code != sample.want {
			t.Fatalf("%s: %d %s", sample.id, response.Code, response.Body.String())
		}
		if sample.want == 200 {
			var result contract.ProductDownload
			if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil || result.FileName != "app.zip" {
				t.Fatalf("download = %+v, %v", result, err)
			}
		}
	}
	if auth.calls != 3 || manifest.reads != 2 {
		t.Fatalf("authorize = %d, manifest reads = %d", auth.calls, manifest.reads)
	}
}
