package services

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"net/mail"
	"strings"
	"time"
	"unicode"

	"github.com/go-redis/redis/v8"
	"golang.org/x/crypto/bcrypt"

	"sitehub/common/config"
	errorx "sitehub/common/errorx"
	"sitehub/common/infrastructure/smtp"
	"sitehub/common/pkg/dberror"
	rbacdomain "sitehub/service/rbac/domain"
	rbacrepository "sitehub/service/rbac/model"
)

const MemberSessionTTL = 7 * 24 * time.Hour

type MemberMailSender func(email, subject, html string) error
type MemberService struct {
	repo   rbacrepository.MemberRepository
	redis  *redis.Client
	secret []byte
	send   MemberMailSender
}

func NewMemberService(repo rbacrepository.MemberRepository, cache *redis.Client, cfg *config.Config) *MemberService {
	return NewMemberServiceWithSender(repo, cache, cfg.Auth.JWTSecret, smtp.SendEmail)
}

// 显式注入邮件发送器，让测试验证邮件内容而不向真实邮箱发送。
func NewMemberServiceWithSender(repo rbacrepository.MemberRepository, cache *redis.Client, secret string, send MemberMailSender) *MemberService {
	return &MemberService{repo, cache, []byte(secret), send}
}
func normalizeMemberEmail(email string) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email || len(email) > 254 || !strings.Contains(email[strings.LastIndex(email, "@")+1:], ".") {
		return "", errorx.Invalid("请输入有效的邮箱地址")
	}
	for _, c := range email {
		if c > 127 {
			return "", errorx.Invalid("请使用标准邮箱地址")
		}
	}
	return email, nil
}
func memberPasswordValid(password string) bool {
	if len(password) < 8 || len(password) > 72 {
		return false
	}
	var letter, number bool
	for _, c := range password {
		letter = letter || unicode.IsLetter(c)
		number = number || unicode.IsDigit(c)
	}
	return letter && number
}
func memberDigest(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
func (s *MemberService) codeHash(email, code string) string {
	h := hmac.New(sha256.New, s.secret)
	h.Write([]byte("product-register:" + email + ":" + code))
	return hex.EncodeToString(h.Sum(nil))
}

// Redis 只负责短期频控；验证码消费和账号创建由同一个 SQL 事务提交。
var memberLimitScript = redis.NewScript(`local n=redis.call('INCR',KEYS[1]); if n==1 then redis.call('EXPIRE',KEYS[1],ARGV[1]) end; return n`)

func (s *MemberService) limit(ctx context.Context, key string, maximum int64, ttl time.Duration) error {
	n, err := memberLimitScript.Run(ctx, s.redis, []string{"product-member:limit:" + memberDigest(key)}, int64(ttl.Seconds())).Int64()
	if err != nil {
		return fmt.Errorf("member rate limit: %w", err)
	}
	if n > maximum {
		return rbacdomain.ErrMemberRateLimit
	}
	return nil
}
func (s *MemberService) SendRegistrationCode(ctx context.Context, email, ip string) error {
	email, err := normalizeMemberEmail(email)
	if err != nil {
		return err
	}
	if len(s.secret) < 16 {
		return errors.New("member verification secret is not configured")
	}
	if err = s.limit(ctx, "send-ip:"+ip, 20, time.Hour); err != nil {
		return err
	}
	if err = s.limit(ctx, "send-minute:"+email, 1, time.Minute); err != nil {
		return err
	}
	if err = s.limit(ctx, "send-hour:"+email, 5, time.Hour); err != nil {
		return err
	}
	number, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return err
	}
	code := fmt.Sprintf("%06d", number.Int64())
	v := rbacdomain.MemberVerification{Email: email, CodeHash: s.codeHash(email, code), ExpiresAt: time.Now().Add(10 * time.Minute)}
	if err = s.repo.WithContext(ctx).PutVerification(v); err != nil {
		return err
	}
	// SMTP 在数据库事务外执行。失败时只撤销本次验证码，不能删掉后来的重发。
	err = s.send(email, "Deploy 官网 · 邮箱注册验证码", fmt.Sprintf("<!DOCTYPE html><html><body><p>你正在注册 Deploy 官网账号。</p><p>验证码：<strong>%s</strong></p><p>10 分钟内有效。请勿向他人透露验证码；如非本人操作，请忽略本邮件。</p></body></html>", code))
	if err != nil {
		cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		return errors.Join(err, s.repo.WithContext(cleanup).DeleteVerification(email, v.CodeHash))
	}
	return nil
}
func (s *MemberService) Register(ctx context.Context, in rbacdomain.RegisterMemberInput, ip string) (rbacdomain.MemberProfile, error) {
	var empty rbacdomain.MemberProfile
	email, err := normalizeMemberEmail(in.Email)
	if err != nil {
		return empty, err
	}
	if !memberPasswordValid(in.Password) {
		return empty, errorx.Invalid("密码需为 8–72 字节，并包含字母和数字")
	}
	if len(in.Code) != 6 || strings.Trim(in.Code, "0123456789") != "" {
		return empty, errorx.Invalid("请输入 6 位邮箱验证码")
	}
	if err = s.limit(ctx, "register-ip:"+ip, 30, 15*time.Minute); err != nil {
		return empty, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return empty, err
	}
	member := rbacdomain.Member{Email: email, PasswordHash: string(hash), CreatedAt: time.Now()}
	var rejection error
	err = s.repo.WithContext(ctx).Transaction(func(tx rbacrepository.MemberRepository) error {
		v, e := tx.Verification(email, true)
		if errors.Is(e, rbacdomain.ErrMemberMissing) {
			rejection = errorx.Invalid("验证码错误或已过期，请重新获取")
			return nil
		}
		if e != nil {
			return e
		}
		// 获取行锁后重新判断过期时间和尝试次数，错误次数必须提交，不能随业务拒绝回滚。
		if !v.ExpiresAt.After(time.Now()) || v.Attempts >= 5 {
			rejection = errorx.Invalid("验证码错误或已过期，请重新获取")
			return nil
		}
		if subtle.ConstantTimeCompare([]byte(v.CodeHash), []byte(s.codeHash(email, in.Code))) != 1 {
			v.Attempts++
			rejection = errorx.Invalid("验证码错误或已过期，请重新获取")
			return tx.PutVerification(v)
		}
		if _, e = tx.MemberByEmail(email); e == nil {
			rejection = errorx.Conflicting("该邮箱已注册，请前往登录")
			return tx.DeleteVerification(email, v.CodeHash)
		} else if !errors.Is(e, rbacdomain.ErrMemberMissing) {
			return e
		}
		if e = tx.CreateMember(&member); e != nil {
			return e
		}
		return tx.DeleteVerification(email, v.CodeHash)
	})
	if dberror.IsDuplicate(err) {
		return empty, errorx.Conflicting("该邮箱已注册，请前往登录")
	}
	if err != nil {
		return empty, err
	}
	if rejection != nil {
		return empty, rejection
	}
	return member.Profile(), nil
}

// 未知邮箱也执行 bcrypt，避免用明显的时间差枚举账号。
var memberDummyHash, _ = bcrypt.GenerateFromPassword([]byte("dummy-member-password-012345"), bcrypt.DefaultCost)

func (s *MemberService) Login(ctx context.Context, in rbacdomain.LoginMemberInput, ip string) (rbacdomain.MemberProfile, string, error) {
	var empty rbacdomain.MemberProfile
	email, err := normalizeMemberEmail(in.Email)
	if err != nil {
		return empty, "", rbacdomain.ErrMemberCredentials
	}
	if len(in.Password) == 0 || len(in.Password) > 72 {
		return empty, "", rbacdomain.ErrMemberCredentials
	}
	if err = s.limit(ctx, "login-ip:"+ip, 30, 15*time.Minute); err != nil {
		return empty, "", err
	}
	if err = s.limit(ctx, "login-email:"+email, 10, 15*time.Minute); err != nil {
		return empty, "", err
	}
	repo := s.repo.WithContext(ctx)
	member, err := repo.MemberByEmail(email)
	if errors.Is(err, rbacdomain.ErrMemberMissing) {
		bcrypt.CompareHashAndPassword(memberDummyHash, []byte(in.Password))
		return empty, "", rbacdomain.ErrMemberCredentials
	}
	if err != nil {
		return empty, "", err
	}
	if bcrypt.CompareHashAndPassword([]byte(member.PasswordHash), []byte(in.Password)) != nil {
		return empty, "", rbacdomain.ErrMemberCredentials
	}
	secret := make([]byte, 32)
	if _, err = rand.Read(secret); err != nil {
		return empty, "", err
	}
	token := hex.EncodeToString(secret)
	err = repo.Transaction(func(tx rbacrepository.MemberRepository) error {
		if e := tx.DeleteExpiredSessions(member.ID, time.Now()); e != nil {
			return e
		}
		return tx.CreateSession(rbacdomain.MemberSession{TokenHash: memberDigest(token), MemberID: member.ID, ExpiresAt: time.Now().Add(MemberSessionTTL)})
	})
	if err != nil {
		return empty, "", err
	}
	return member.Profile(), token, nil
}
func (s *MemberService) Profile(ctx context.Context, token string) (rbacdomain.MemberProfile, error) {
	if len(token) != 64 {
		return rbacdomain.MemberProfile{}, rbacdomain.ErrMemberSession
	}
	member, err := s.repo.WithContext(ctx).SessionMember(memberDigest(token), time.Now())
	if errors.Is(err, rbacdomain.ErrMemberMissing) {
		err = rbacdomain.ErrMemberSession
	}
	return member.Profile(), err
}
func (s *MemberService) Logout(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	return s.repo.WithContext(ctx).DeleteSession(memberDigest(token))
}
