package services

import (
	"context"
	"errors"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"sitehub/service/rbac/domain"
	"sitehub/service/rbac/model/mysql"
)

// Reset codes must never accept a registration challenge or leave previously issued sessions valid.
func TestMemberRecoveryRevokesSessions(t *testing.T) {
	ctx := context.Background()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "members.db")), &gorm.Config{TranslateError: true})
	check := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	check(err)
	pool, err := db.DB()
	check(err)
	t.Cleanup(func() { pool.Close() })
	check(db.AutoMigrate(&domain.Member{}, &domain.MemberSession{}, &domain.MemberVerification{}))
	server := miniredis.RunT(t)
	cache := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { cache.Close() })
	var code string
	capture := func(email, subject, html string) error {
		matches := regexp.MustCompile(`<strong>([0-9]{6})</strong>`).FindStringSubmatch(html)
		if len(matches) != 2 {
			t.Fatal("missing verification code")
		}
		code = matches[1]
		return nil
	}
	s := NewMemberServiceWithSender(mysql.NewMemberRepository(db), cache, "isolated-verification-secret", capture)
	const email = "reader@example.test"
	check(s.SendResetCode(ctx, email, "test"))
	if code != "" {
		t.Fatal("delivered reset mail before account creation")
	}
	server.FastForward(time.Minute)
	check(s.SendRegistrationCode(ctx, email, "test"))
	registration := domain.RegisterMemberInput{Email: email, Password: "Reader1234", Code: code, Username: "Reader1", Nickname: "山窗"}
	reset := domain.ResetMemberPasswordInput{Email: email, Password: "Changed1234", Code: code}
	if err := s.ResetPassword(ctx, reset, "test"); err == nil {
		t.Fatal("accepted registration code for reset")
	}
	_, err = s.Register(ctx, registration, "test")
	check(err)
	credentials := domain.LoginMemberInput{Account: "reader1", Password: registration.Password}
	_, first, err := s.Login(ctx, credentials, "test")
	check(err)
	_, second, err := s.Login(ctx, credentials, "test")
	check(err)
	changed, err := s.UpdateProfile(ctx, first, domain.UpdateMemberProfileInput{Username: "Reader1", Nickname: "林舟", Avatar: "https://example.test/avatar.png"})
	check(err)
	restored, err := s.Profile(ctx, second)
	check(err)
	if restored.Nickname != "林舟" || restored.Avatar != changed.Avatar {
		t.Fatal("profile changes were not persisted")
	}
	server.FastForward(time.Minute)
	check(s.SendResetCode(ctx, email, "test"))
	reset.Code = code
	check(s.ResetPassword(ctx, reset, "test"))
	if err := s.ResetPassword(ctx, reset, "test"); err == nil {
		t.Fatal("reset code was reusable")
	}
	for _, token := range []string{first, second} {
		if _, err := s.Profile(ctx, token); !errors.Is(err, domain.ErrMemberSession) {
			t.Fatalf("revoked session: %v", err)
		}
	}
	if _, _, err := s.Login(ctx, credentials, "test"); !errors.Is(err, domain.ErrMemberCredentials) {
		t.Fatalf("old password: %v", err)
	}
	_, _, err = s.Login(ctx, domain.LoginMemberInput{Email: email, Password: reset.Password}, "test")
	check(err)
}
