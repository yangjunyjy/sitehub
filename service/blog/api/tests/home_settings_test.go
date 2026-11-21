package api_test

import (
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
	"net/http/httptest"
	rbacpb "sitehub/service/rbac/rpc/pb/rbac"
	"strings"
	"testing"
)

func TestHomeConfigurationPreservesExplicitFalse(t *testing.T) {
	handler, mock, _ := unifiedAPI(t)
	mock.ExpectQuery("SELECT .*FROM `author_profiles`").WillReturnRows(sqlmock.NewRows([]string{"id", "home_animation", "home_verse"}).AddRow(1, false, "山色入窗"))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest("GET", "/v1/blog/about", nil))
	body := recorder.Body.String()
	if recorder.Code != 200 || !strings.Contains(body, `"homeAnimation":false`) || !strings.Contains(body, `"homeVerse":"山色入窗"`) {
		t.Fatal(body)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

type homeSettingsAuthorizer struct{ rbacpb.RbacClient }

func (homeSettingsAuthorizer) Authorize(_ context.Context, req *rbacpb.AuthorizeRequest, _ ...grpc.CallOption) (*rbacpb.Principal, error) {
	if req.Token != "admin-token" || req.Permission != "system:settings:edit" || req.Audience != "admin" || !req.Secure {
		return nil, status.Error(codes.PermissionDenied, "forbidden")
	}
	return &rbacpb.Principal{}, nil
}

func TestSaveHomeConfiguration(t *testing.T) {
	handler, mock, _ := unifiedAPI(t, homeSettingsAuthorizer{})
	for _, sample := range []struct {
		token, verse string
		code         int
	}{
		{"", "山色入窗", http.StatusForbidden},
		{"admin-token", "   ", http.StatusBadRequest},
		{"admin-token", "山色入窗", http.StatusOK},
	} {
		if sample.code == http.StatusOK {
			mock.ExpectBegin()
			mock.ExpectExec("INSERT INTO `author_profiles` .* ON DUPLICATE KEY UPDATE `home_animation`=VALUES.*`home_verse`=VALUES").WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectCommit()
		}
		req := httptest.NewRequest(http.MethodPost, "/v1/blog/home/settings", strings.NewReader(`{"homeAnimation":false,"homeVerse":"`+sample.verse+`"}`))
		req.Header.Set("Content-Type", "application/json")
		if sample.token != "" {
			req.Header.Set("Authorization", "Bearer "+sample.token)
		}
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, req)
		if recorder.Code != sample.code {
			t.Fatalf("want %d, got %d: %s", sample.code, recorder.Code, recorder.Body.String())
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
