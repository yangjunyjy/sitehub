package handler

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

func TestCommunityHTTPRoutes(t *testing.T) {
	handler, mock, _ := unifiedAPI(t)
	mock.ExpectQuery("SELECT .*FROM `friend_links` WHERE status = .*ORDER BY id DESC").WithArgs("approved").WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "Friends"))
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `friend_links`").WithArgs("Friend", "https://example.org", "notes", "", "pending", sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(2, 1))
	mock.ExpectCommit()
	mock.ExpectQuery("SELECT count.*FROM `guest_messages`").WithArgs("approved", 0).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(40))
	mock.ExpectQuery("SELECT .*FROM `guest_messages` WHERE status = .*ORDER BY id DESC LIMIT .* OFFSET").WithArgs("approved", 0, 40, 40).WillReturnRows(sqlmock.NewRows([]string{"id", "content"}))
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `guest_messages`").WithArgs(0, "visitor", "hello", "pending", sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(3, 1))
	mock.ExpectCommit()
	mock.ExpectQuery("SELECT .*FROM `author_profiles`").WillReturnRows(sqlmock.NewRows([]string{"id", "skills"}).AddRow(1, `["Go"]`))
	for _, sample := range []struct{ method, path, body, expected string }{
		{"GET", "/v1/blog/friends", "", `"name":"Friends"`},
		{"POST", "/v1/blog/friends", `{"name":"Friend","url":"https://example.org","description":"notes"}`, `"status":"pending"`},
		{"GET", "/v1/blog/messages?page=2&pageSize=40", "", `"items":[]`},
		{"POST", "/v1/blog/messages", `{"nickname":"visitor","content":"hello"}`, `"status":"pending"`},
		{"GET", "/v1/blog/about", "", `"skills":["Go"]`},
	} {
		request := httptest.NewRequest(sample.method, sample.path, strings.NewReader(sample.body))
		if sample.body != "" {
			request.Header.Set("Content-Type", "application/json")
		}
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		if recorder.Code != 200 || !strings.Contains(recorder.Body.String(), sample.expected) {
			t.Errorf("%s %s: %d %s", sample.method, sample.path, recorder.Code, recorder.Body.String())
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

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

func TestCommunityRejectsInvalidPublicInput(t *testing.T) {
	handler, _, _ := unifiedAPI(t)
	for _, test := range []struct{ method, path, body string }{
		{"POST", "/v1/blog/friends", `{"name":"test","description":"test","url":"javascript:alert(1)"}`},
		{"POST", "/v1/blog/friends", `{"name":"test","description":"test","url":"/relative"}`},
		{"POST", "/v1/blog/friends", `{"name":"test","description":"test","url":"https://name:pass@example.org"}`},
		{"POST", "/v1/blog/messages", `{"content":" "}`},
		{"POST", "/v1/blog/messages", `{"content":"` + strings.Repeat("a", 1001) + `"}`},
		{"GET", "/v1/blog/messages?page=-1", ""},
	} {
		req := httptest.NewRequest(test.method, test.path, strings.NewReader(test.body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		handler.ServeHTTP(resp, req)
		if resp.Code != http.StatusBadRequest {
			t.Errorf("%s %s: %d", test.method, test.path, resp.Code)
		}
	}
}
