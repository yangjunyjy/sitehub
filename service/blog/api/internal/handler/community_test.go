package handler

import (
	"github.com/DATA-DOG/go-sqlmock"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCommunityHTTPRoutes(t *testing.T) {
	handler, mock, _ := unifiedAPI(t)
	mock.ExpectQuery("SELECT .*FROM `friend_links` WHERE status = .*ORDER BY id DESC").WithArgs("approved").WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "Friends"))
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `friend_links`").WithArgs("Friend", "https://example.org", "notes", "", "pending", sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(2, 1))
	mock.ExpectCommit()
	mock.ExpectQuery("SELECT .*FROM `guest_messages` WHERE status = .*ORDER BY id DESC LIMIT .* OFFSET").WithArgs("approved", 40, 40).WillReturnRows(sqlmock.NewRows([]string{"id", "content"}))
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `guest_messages`").WithArgs("visitor", "hello", "pending", sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(3, 1))
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
