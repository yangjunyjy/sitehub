package servicehttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sitehub/common/errorx"
	"sitehub/common/rpcx"
)

func TestDownloadForwardsVerifiedRequestIdentity(t *testing.T) {
	key := strings.Repeat("k", 32)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !Authorized(r, key) || r.Header.Get("Authorization") != "Bearer member" {
			t.Error("missing service or member credentials")
		}
		Write(w, nil, &errorx.PublicError{Kind: errorx.NotFound, Message: "unpublished"})
	}))
	defer upstream.Close()
	req := httptest.NewRequest("GET", "/download", nil)
	req.Header.Set("Authorization", "Bearer member")
	rpcx.HTTPContext(func(_ http.ResponseWriter, request *http.Request) {
		err := Do(request.Context(), upstream.URL, key, "GET", "/internal/product/download?id=1", nil, nil)
		if status.Code(err) != codes.NotFound || status.Convert(err).Message() != "unpublished" {
			t.Fatalf("lost download status: %v", err)
		}
	})(httptest.NewRecorder(), req)
}

func TestRedirectDoesNotForwardServiceCredentials(t *testing.T) {
	destination := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { t.Error("followed internal service redirect") }))
	defer destination.Close()
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, destination.URL, http.StatusTemporaryRedirect)
	}))
	defer origin.Close()
	if err := Do(context.Background(), origin.URL, strings.Repeat("s", 32), "GET", "/internal/blog/storage", nil, nil); status.Code(err) != codes.Unavailable {
		t.Fatalf("unexpected redirect result: %v", err)
	}
}
