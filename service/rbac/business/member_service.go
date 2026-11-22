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
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

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
func (s *MemberService) codeHash(purpose, email, code string) string {
	h := hmac.New(sha256.New, s.secret)
	h.Write([]byte("product-" + purpose + ":" + email + ":" + code))
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
	return s.sendCode(ctx, email, ip, "register")
}
func (s *MemberService) SendResetCode(ctx context.Context, email, ip string) error {
	return s.sendCode(ctx, email, ip, "reset")
}
func (s *MemberService) sendCode(ctx context.Context, email, ip, purpose string) error {
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
	action := "注册会员账号"
	if purpose == "reset" {
		_, err := s.repo.WithContext(ctx).MemberByEmail(email)
		if errors.Is(err, rbacdomain.ErrMemberMissing) {
			return nil
		}
		if err != nil {
			return err
		}
		action = "重置会员密码"
	}
	number, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return err
	}
	code := fmt.Sprintf("%06d", number.Int64())
	v := rbacdomain.MemberVerification{Email: email, CodeHash: s.codeHash(purpose, email, code), ExpiresAt: time.Now().Add(10 * time.Minute)}
	if err = s.repo.WithContext(ctx).PutVerification(v); err != nil {
		return err
	}
	// SMTP 在数据库事务外执行。失败时只撤销本次验证码，不能删掉后来的重发。
	err = s.send(email, action+" · 邮箱验证码", fmt.Sprintf("<!DOCTYPE html><html><body><p>你正在%s。</p><p>验证码：<strong>%s</strong></p><p>10 分钟内有效。请勿向他人透露验证码；如非本人操作，请忽略本邮件。</p></body></html>", action, code))
	if err != nil {
		cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		return errors.Join(err, s.repo.WithContext(cleanup).DeleteVerification(email, v.CodeHash))
	}
	return nil
}

var memberUsername = regexp.MustCompile(`^[a-z0-9_]{3,40}$`)

func normalizeProfile(in rbacdomain.UpdateMemberProfileInput) (rbacdomain.UpdateMemberProfileInput, error) {
	in.Username = strings.ToLower(strings.TrimSpace(in.Username))
	in.Nickname = strings.TrimSpace(in.Nickname)
	in.Avatar = strings.TrimSpace(in.Avatar)
	if !memberUsername.MatchString(in.Username) {
		return in, errorx.Invalid("用户名需为 3–40 位字母、数字或下划线")
	}
	if in.Nickname == "" || utf8.RuneCountInString(in.Nickname) > 40 || strings.IndexFunc(in.Nickname, unicode.IsControl) >= 0 {
		return in, errorx.Invalid("昵称需为 1–40 个可显示字符")
	}
	if in.Avatar != "" {
		u, err := url.ParseRequestURI(in.Avatar)
		if len(in.Avatar) > 1024 || err != nil || u.Hostname() == "" || u.User != nil || (u.Scheme != "https" && u.Scheme != "http") {
			return in, errorx.Invalid("请输入有效的 HTTP 或 HTTPS 头像图片地址")
		}
	}
	return in, nil
}

func (s *MemberService) Register(ctx context.Context, in rbacdomain.RegisterMemberInput, ip string) (rbacdomain.MemberProfile, error) {
	profile, err := normalizeProfile(rbacdomain.UpdateMemberProfileInput{Username: in.Username, Nickname: in.Nickname})
	if err != nil {
		return rbacdomain.MemberProfile{}, err
	}
	var member rbacdomain.Member
	err = s.withVerifiedPassword(ctx, "register", ip, in.Email, in.Code, in.Password, func(tx rbacrepository.MemberRepository, email, hash string) error {
		member = rbacdomain.Member{Email: email, PasswordHash: hash, Username: &profile.Username, Nickname: profile.Nickname, CreatedAt: time.Now()}
		return tx.CreateMember(&member)
	})
	if dberror.IsDuplicate(err) {
		err = errorx.Conflicting("用户名或邮箱已被使用")
	}
	if err != nil {
		return rbacdomain.MemberProfile{}, err
	}
	return member.Profile(), nil
}

func (s *MemberService) ResetPassword(ctx context.Context, in rbacdomain.ResetMemberPasswordInput, ip string) error {
	return s.withVerifiedPassword(ctx, "reset", ip, in.Email, in.Code, in.Password, func(tx rbacrepository.MemberRepository, email, hash string) error {
		m, err := tx.MemberByEmail(email)
		if err != nil {
			return err
		}
		if _, err = tx.LockMember(m.ID); err != nil {
			return err
		}
		return tx.ResetPassword(m.ID, hash)
	})
}

// 锁定验证码后校验用途、时效与次数，验证码消费和密码写入一起提交。
func (s *MemberService) withVerifiedPassword(ctx context.Context, purpose, ip, address, code, password string, apply func(rbacrepository.MemberRepository, string, string) error) error {
	email, err := normalizeMemberEmail(address)
	if err != nil {
		return err
	}
	if !memberPasswordValid(password) {
		return errorx.Invalid("密码需为 8–72 字节，并包含字母和数字")
	}
	if len(code) != 6 || strings.Trim(code, "0123456789") != "" {
		return errorx.Invalid("请输入 6 位邮箱验证码")
	}
	if err = s.limit(ctx, purpose+"-ip:"+ip, 30, 15*time.Minute); err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	var rejected bool
	err = s.repo.WithContext(ctx).Transaction(func(tx rbacrepository.MemberRepository) error {
		v, err := tx.Verification(email, true)
		if errors.Is(err, rbacdomain.ErrMemberMissing) {
			rejected = true
			return nil
		}
		if err != nil {
			return err
		}
		if !v.ExpiresAt.After(time.Now()) || v.Attempts >= 5 {
			rejected = true
			return nil
		}
		if subtle.ConstantTimeCompare([]byte(v.CodeHash), []byte(s.codeHash(purpose, email, code))) != 1 {
			rejected = true
			v.Attempts++
			return tx.PutVerification(v)
		}
		if err = apply(tx, email, string(hash)); err != nil {
			return err
		}
		return tx.DeleteVerification(email, v.CodeHash)
	})
	if err != nil {
		return err
	}
	if rejected {
		return errorx.Invalid("验证码错误或已过期，请重新获取")
	}
	return nil
}

func (s *MemberService) UpdateProfile(ctx context.Context, token string, input rbacdomain.UpdateMemberProfileInput) (rbacdomain.MemberProfile, error) {
	p, err := normalizeProfile(input)
	if err != nil {
		return rbacdomain.MemberProfile{}, err
	}
	var result rbacdomain.MemberProfile
	err = s.repo.WithContext(ctx).Transaction(func(tx rbacrepository.MemberRepository) error {
		m, err := tx.SessionMember(memberDigest(token), time.Now())
		if errors.Is(err, rbacdomain.ErrMemberMissing) {
			return rbacdomain.ErrMemberSession
		}
		if err != nil {
			return err
		}
		if err = tx.SaveProfile(m.ID, p); err != nil {
			return err
		}
		m.Username, m.Nickname, m.Avatar = &p.Username, p.Nickname, p.Avatar
		result = m.Profile()
		return nil
	})
	if dberror.IsDuplicate(err) {
		return result, errorx.Conflicting("该用户名已被使用")
	}
	return result, err
}

// 未知邮箱也执行 bcrypt，避免用明显的时间差枚举账号。
var memberDummyHash, _ = bcrypt.GenerateFromPassword([]byte("dummy-member-password-012345"), bcrypt.DefaultCost)

func (s *MemberService) Login(ctx context.Context, in rbacdomain.LoginMemberInput, ip string) (rbacdomain.MemberProfile, string, error) {
	var empty rbacdomain.MemberProfile
	account := strings.ToLower(strings.TrimSpace(in.Account))
	if account == "" {
		account = strings.ToLower(strings.TrimSpace(in.Email))
	}
	repo := s.repo.WithContext(ctx)
	lookup := repo.MemberByUsername
	if strings.Contains(account, "@") {
		if _, err := normalizeMemberEmail(account); err != nil {
			return empty, "", rbacdomain.ErrMemberCredentials
		}
		lookup = repo.MemberByEmail
	} else if !memberUsername.MatchString(account) {
		return empty, "", rbacdomain.ErrMemberCredentials
	}
	if len(in.Password) == 0 || len(in.Password) > 72 {
		return empty, "", rbacdomain.ErrMemberCredentials
	}
	if err := s.limit(ctx, "login-ip:"+ip, 30, 15*time.Minute); err != nil {
		return empty, "", err
	}
	if err := s.limit(ctx, "login-account:"+account, 10, 15*time.Minute); err != nil {
		return empty, "", err
	}
	member, err := lookup(account)
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
		// 与密码重置竞争时，以持锁后的密码版本决定能否签发会话。
		current, e := tx.LockMember(member.ID)
		if e != nil {
			return e
		}
		if current.PasswordHash != member.PasswordHash {
			return rbacdomain.ErrMemberCredentials
		}
		member = current
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
