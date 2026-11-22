package console

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"os"
	"strings"
	"time"

	"sitehub/common/rpcx"
)

func (c *console) setSession(w http.ResponseWriter, token string, age int) {
	path := c.CookiePath
	if path == "" {
		path = "/api/"
	}
	http.SetCookie(w, &http.Cookie{Name: sessionCookie, Value: token, Path: path, HttpOnly: true, SameSite: http.SameSiteStrictMode, Secure: strings.HasPrefix(c.Origin, "https://"), MaxAge: age})
}

func (c *console) login(w http.ResponseWriter, r *http.Request) {
	contentType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || contentType != "application/json" {
		problem(w, http.StatusUnsupportedMediaType, "请提交 JSON")
		return
	}
	var in struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<10))
	dec.DisallowUnknownFields()
	if dec.Decode(&in) != nil || dec.Decode(new(any)) != io.EOF {
		problem(w, http.StatusBadRequest, "登录请求格式有误")
		return
	}
	a, err := c.identity.load(in.Username)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		problem(w, http.StatusServiceUnavailable, "无法读取控制台账号")
		return
	}
	if !c.identity.checkPassword(w, rpcx.HTTPValue(r.Context(), "x-client-ip"), a, in.Password) {
		return
	}
	c.identity.revoke(r)
	token, ok := c.identity.issue(a)
	if !ok {
		problem(w, http.StatusServiceUnavailable, "控制台会话已满，请稍后重试")
		return
	}
	c.setSession(w, token, int(sessionLifetime/time.Second))
	w.WriteHeader(http.StatusNoContent)
}

func (c *console) logout(w http.ResponseWriter, r *http.Request) {
	c.identity.revoke(r)
	c.setSession(w, "", -1)
	w.WriteHeader(http.StatusNoContent)
}
