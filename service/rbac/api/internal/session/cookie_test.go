package session

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCookiesAreIssuedByRBACAndClearedIndependently(t *testing.T) {
	for _, audience := range []string{"admin", "site"} {
		w := httptest.NewRecorder()
		h := Middleware(Config{Secure: true})(func(w http.ResponseWriter, r *http.Request) {
			Issue(r.Context(), audience, "opaque-token", time.Now().Add(time.Hour).Unix())
		})
		h(w, httptest.NewRequest("POST", "/login", nil))
		cookies := w.Result().Cookies()
		if len(cookies) != 1 || cookies[0].Name != "sitehub_"+audience+"_session" || !cookies[0].Secure || !cookies[0].HttpOnly || cookies[0].Domain != "" || cookies[0].SameSite != http.SameSiteLaxMode {
			t.Fatalf("unsafe cookie: %+v", cookies)
		}
		w = httptest.NewRecorder()
		Middleware(Config{})(func(w http.ResponseWriter, r *http.Request) { Clear(r.Context(), audience) })(w, httptest.NewRequest("POST", "/logout", nil))
		if w.Result().Cookies()[0].MaxAge != -1 {
			t.Fatal("logout did not delete cookie")
		}
	}
}
