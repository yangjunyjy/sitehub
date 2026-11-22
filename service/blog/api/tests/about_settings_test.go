package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"sitehub/common/apiresponse"
	"sitehub/common/rpcx"
	blog "sitehub/service/blog"
	"sitehub/service/blog/api/internal/handler/settings"
	"sitehub/service/blog/api/internal/svc"
	"sitehub/service/blog/model/entity"
	"sitehub/service/blog/model/mysql"
	pb "sitehub/service/rbac/rpc/pb/rbac"
)

type aboutAuthorizer struct{ pb.RbacClient }

func (aboutAuthorizer) Authorize(_ context.Context, req *pb.AuthorizeRequest, _ ...grpc.CallOption) (*pb.Principal, error) {
	if req.Token != "editor" || req.Permission != "blog:settings:edit" || req.Audience != "admin" || !req.Secure {
		return nil, status.Error(codes.PermissionDenied, "forbidden")
	}
	return &pb.Principal{}, nil
}

// Both views use author_profiles.id=1; either editor must be able to save first without losing the other editor's fields.
func TestAboutPageSettingsRoundTrip(t *testing.T) {
	apiresponse.Install()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "about.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	pool, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	if err = db.AutoMigrate(&entity.AuthorProfile{}); err != nil {
		t.Fatal(err)
	}
	repo := mysql.NewCommunityRepository(db)
	ctx := &svc.ServiceContext{Core: &blog.Module{Community: repo}, Rbac: aboutAuthorizer{}}
	save := rpcx.HTTPContext(settings.SaveAboutSettingsHandler(ctx))
	load := rpcx.HTTPContext(settings.GetAboutHandler(ctx))
	post := func(token, body string, want int) {
		t.Helper()
		r := httptest.NewRequest("POST", "/v1/blog/about/settings", strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		if token != "" {
			r.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()
		save.ServeHTTP(w, r)
		if w.Code != want {
			t.Fatalf("status %d: %s", w.Code, w.Body.String())
		}
	}
	read := func() map[string]any {
		t.Helper()
		w := httptest.NewRecorder()
		load.ServeHTTP(w, httptest.NewRequest("GET", "/v1/blog/about", nil))
		var reply struct{ Data map[string]any }
		if w.Code != http.StatusOK || json.Unmarshal(w.Body.Bytes(), &reply) != nil {
			t.Fatal(w.Body.String())
		}
		return reply.Data
	}
	post("", "{}", 403)
	post("editor", `{"avatar":"//other.example/avatar"}`, 400)
	post("editor", `{"socials":[{"label":"Profile","url":"invalid"}]}`, 400)
	post("editor", `{"name":"林舟","skills":["Go","摄影"],"socials":[{"label":"主页","url":"https://example.org/"}]}`, 200)
	if err = repo.SaveHome(context.Background(), false, "山风"); err != nil {
		t.Fatal(err)
	}
	profile := read()
	if profile["name"] != "林舟" || len(profile["skills"].([]any)) != 2 || len(profile["socials"].([]any)) != 1 {
		t.Fatal(profile)
	}
	post("editor", `{"name":"木木","skills":[],"socials":[]}`, 200)
	profile = read()
	if profile["homeVerse"] != "山风" || profile["homeAnimation"] != false || profile["name"] != "木木" || len(profile["skills"].([]any)) != 0 || len(profile["socials"].([]any)) != 0 {
		t.Fatal(profile)
	}
}
