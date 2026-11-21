package api_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/router"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"sitehub/common/apiresponse"
	"sitehub/common/rpcx"
	"sitehub/service/product"
	"sitehub/service/product/api/internal/handler"
	"sitehub/service/product/api/internal/svc"
	"sitehub/service/product/business"
	"sitehub/service/product/contract"
	"sitehub/service/product/model/mysql"
	"sitehub/service/product/model/schema"
	rbac "sitehub/service/rbac/rpc/pb/rbac"
)

type manualEditorAuth struct{ rbac.RbacClient }

func (manualEditorAuth) Authorize(_ context.Context, req *rbac.AuthorizeRequest, _ ...grpc.CallOption) (*rbac.Principal, error) {
	if req.Token != "editor" || req.Permission != "website:docs:edit" || req.Audience != "admin" {
		return nil, status.Error(codes.PermissionDenied, "文档编辑权限不足")
	}
	if req.Secure && req.Totp != "test-totp" {
		return nil, status.Error(codes.PermissionDenied, "需要二次验证")
	}
	return &rbac.Principal{Id: 19}, nil
}
func TestRegisteredManualWorkspaceRoutes(t *testing.T) {
	apiresponse.Install()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "workspace.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	pool, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { pool.Close() })
	if err = schema.Apply(db); err != nil {
		t.Fatal(err)
	}
	server, err := rest.NewServer(rest.RestConf{})
	if err != nil {
		t.Fatal(err)
	}
	handler.RegisterHandlers(server, &svc.ServiceContext{Core: &product.Module{Manuals: services.NewManualService(mysql.NewProductDocsRepository(db))}, Rbac: manualEditorAuth{}})
	mux := router.NewRouter()
	for _, route := range server.Routes() {
		if err = mux.Handle(route.Method, route.Path, rpcx.HTTPContext(route.Handler)); err != nil {
			t.Fatal(err)
		}
	}
	request := func(method, path, body, token string, want int) []byte {
		t.Helper()
		req := httptest.NewRequest(method, "/v1/product/"+path, strings.NewReader(body))
		if body != "" {
			req.Header.Set("Content-Type", "application/json")
		}
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("X-Totp-Token", "test-totp")
		}
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != want {
			t.Fatalf("%s %s: got %d want %d: %s", method, path, rec.Code, want, rec.Body.String())
		}
		return rec.Body.Bytes()
	}
	const admin = "admin/manuals/"
	request(http.MethodGet, admin+"workspace", "", "", 403)
	request(http.MethodPost, admin+"groups/save", `{"version":"v1","title":"Basics"}`, "editor", 200)
	var workspace struct{ Data contract.ManualWorkspace }
	err = json.Unmarshal(request("GET", admin+"workspace?mode=guide&version=v1", "", "editor", 200), &workspace)
	if err != nil {
		t.Fatal(err)
	}
	if len(workspace.Data.Groups) != 1 {
		t.Fatal(workspace.Data)
	}
	payload := fmt.Sprintf(`{"version":"v1","id":"install","groupId":%d,"slug":"install","title":"Install","content":"approved text","revision":0}`, workspace.Data.Groups[0].ID)
	request("POST", admin+"documents/save", payload, "editor", 200)
	request("GET", "manuals/document?id=install&version=v1", "", "", 404)
	ref := `{"id":"install","version":"v1","revision":1}`
	request("POST", admin+"documents/publish", ref, "", 403)
	request("POST", admin+"documents/publish", ref, "editor", 200)
	request("POST", admin+"documents/publish", ref, "editor", 409)
	result := request("GET", "manuals/document?id=install&version=v1", "", "", 200)
	var published struct{ Data contract.WebsiteDocDocument }
	if err = json.Unmarshal(result, &published); err != nil || published.Data.Content != "approved text" {
		t.Fatalf("publication: %s %v", result, err)
	}
	request("POST", admin+"documents/withdraw", `{"id":"install","version":"v1","revision":2}`, "editor", 200)
	request("GET", "manuals/document?id=install&version=v1", "", "", 404)
}
