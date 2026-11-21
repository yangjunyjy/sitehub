package api_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "sitehub/service/rbac/rpc/pb/rbac"
)

type readerAuthorizer struct{ pb.RbacClient }

func (readerAuthorizer) Authorize(_ context.Context, req *pb.AuthorizeRequest, _ ...grpc.CallOption) (*pb.Principal, error) {
	if req.Token != "reader-token" || req.Audience != "member" {
		return nil, status.Error(codes.Unauthenticated, "请登录")
	}
	return &pb.Principal{Id: 7, Name: "注册读者", Member: true}, nil
}

func TestCommunityWritesRequireMembership(t *testing.T) {
	mux, _, _ := unifiedAPI(t, readerAuthorizer{})
	for _, sample := range []struct{ path, body string }{
		{"/v1/blog/messages", `{"content":"hello"}`}, {"/v1/blog/messages", `{"content":"hi","articleId":1}`},
		{"/v1/blog/friends", `{"name":"test","description":"hi","url":"https://example.org"}`},
		{"/v1/blog/articles/like", `{"articleId":1,"liked":true}`}, {"/v1/blog/articles/favorite", `{"articleId":1,"favorited":true}`},
		{"/v1/blog/favorites/folders", `{"name":"folder"}`},
	} {
		req := httptest.NewRequest("POST", sample.path, strings.NewReader(sample.body))
		req.Header.Set("Content-Type", "application/json")
		res := httptest.NewRecorder()
		mux.ServeHTTP(res, req)
		if res.Code != 401 {
			t.Errorf("%s: %d %s", sample.path, res.Code, res.Body.String())
		}
	}
}

func TestCommunityHTTPRoutes(t *testing.T) {
	handler, mock, _ := unifiedAPI(t, readerAuthorizer{})
	mock.ExpectQuery("SELECT .*FROM `friend_links` WHERE status = .*ORDER BY id DESC").WithArgs("approved").WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "Friends"))
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `friend_links`").WithArgs(7, "Friend", "https://example.org", "notes", "", "pending", sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(2, 1))
	mock.ExpectCommit()
	mock.ExpectQuery("SELECT count.*FROM `guest_messages`").WithArgs("approved", 0).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(40))
	mock.ExpectQuery("SELECT .*FROM `guest_messages` WHERE status = .*ORDER BY id DESC LIMIT .* OFFSET").WithArgs("approved", 0, 40, 40).WillReturnRows(sqlmock.NewRows([]string{"id", "content"}))
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `guest_messages`").WithArgs(7, 0, "注册读者", "hello", "pending", sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(3, 1))
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
			request.Header.Set("Authorization", "Bearer reader-token")
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

func TestCommunityRejectsInvalidPublicInput(t *testing.T) {
	handler, _, _ := unifiedAPI(t, readerAuthorizer{})
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
		req.Header.Set("Authorization", "Bearer reader-token")
		resp := httptest.NewRecorder()
		handler.ServeHTTP(resp, req)
		if resp.Code != http.StatusBadRequest {
			t.Errorf("%s %s: %d", test.method, test.path, resp.Code)
		}
	}
}
