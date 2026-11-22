package console

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	sessionCookie   = "sitehub_gateway_session"
	sessionLifetime = 8 * time.Hour
)

var accountName = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,64}$`)

type account struct {
	Name string `json:"username"`
	Role string `json:"role"`
	Hash string `json:"password_hash"`
}

func (a account) allows(permission string) bool {
	switch permission {
	case readConfig:
		return a.Role == "admin" || a.Role == "publisher" || a.Role == "viewer"
	case publishConfig:
		return a.Role == "admin" || a.Role == "publisher"
	case rollbackConfig:
		return a.Role == "admin"
	}
	return false
}

func accountFile(dir, name string) (string, error) {
	if dir == "" || !accountName.MatchString(name) {
		return "", fmt.Errorf("a state directory and a username of 1-64 ASCII letters, digits, underscores or hyphens are required")
	}
	// Use a digest for uniform case-sensitive identity across Windows and Unix.
	return filepath.Join(dir, "accounts", fmt.Sprintf("%x.json", sha256.Sum256([]byte(name)))), nil
}

// SetAccount creates or replaces one account. Each newly salted password hash also
// invalidates sessions established before a password reset or role change.
func SetAccount(dir, name, role string, password []byte) error {
	path, err := accountFile(dir, name)
	if err != nil {
		return err
	}
	user := account{Name: name, Role: role}
	if !user.allows(readConfig) {
		return fmt.Errorf("role must be admin, publisher or viewer")
	}
	if len(password) < 12 || len(password) > 72 {
		return fmt.Errorf("password must contain 12 to 72 bytes")
	}
	hash, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Hash = string(hash)
	if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".account-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if err = json.NewEncoder(tmp).Encode(user); err == nil {
		err = tmp.Sync()
	}
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}

func RemoveAccount(dir, name string) error {
	path, err := accountFile(dir, name)
	if err != nil {
		return err
	}
	return os.Remove(path)
}

type session struct {
	name, credential string
	expires          time.Time
}

type attempts struct {
	count   int
	expires time.Time
}

type identityStore struct {
	dir       string
	dummyHash []byte
	mu        sync.Mutex
	sessions  map[[32]byte]session
	attempts  map[string]attempts
	checks    chan struct{}
}

func newIdentityStore(dir string) (*identityStore, error) {
	if dir == "" {
		return nil, errors.New("console state directory is required")
	}
	dummy, err := bcrypt.GenerateFromPassword([]byte(rand.Text()), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return &identityStore{dir: dir, dummyHash: dummy, sessions: make(map[[32]byte]session),
		attempts: make(map[string]attempts), checks: make(chan struct{}, 4)}, nil
}

func (s *identityStore) load(name string) (account, error) {
	path, err := accountFile(s.dir, name)
	if err != nil {
		return account{}, os.ErrNotExist
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return account{}, err
	}
	var a account
	if json.Unmarshal(data, &a) != nil || a.Name != name || !a.allows(readConfig) {
		return account{}, errors.New("invalid local console account")
	}
	if _, err := bcrypt.Cost([]byte(a.Hash)); err != nil {
		return account{}, errors.New("invalid local console credential")
	}
	return a, nil
}

// Bound both per-client failures and concurrent password hashes on the proxy process.
func (s *identityStore) reserve(ip string, now time.Time) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	a := s.attempts[ip]
	if !now.Before(a.expires) {
		if len(s.attempts) >= 4096 {
			for key, entry := range s.attempts {
				if !now.Before(entry.expires) {
					delete(s.attempts, key)
				}
			}
			if len(s.attempts) >= 4096 {
				return false
			}
		}
		a = attempts{expires: now.Add(5 * time.Minute)}
	}
	if a.count >= 10 {
		return false
	}
	a.count++
	s.attempts[ip] = a
	return true
}

func (s *identityStore) checkPassword(w http.ResponseWriter, ip string, a account, password string) bool {
	select {
	case s.checks <- struct{}{}:
		defer func() { <-s.checks }()
	default:
		problem(w, http.StatusTooManyRequests, "登录请求繁忙，请稍后重试")
		return false
	}
	if !s.reserve(ip, time.Now()) {
		problem(w, http.StatusTooManyRequests, "密码尝试次数过多，请稍后重试")
		return false
	}
	hash := s.dummyHash
	if a.Hash != "" {
		hash = []byte(a.Hash)
	}
	if len(password) > 72 || bcrypt.CompareHashAndPassword(hash, []byte(password)) != nil || a.Hash == "" {
		problem(w, http.StatusForbidden, "账号或密码有误")
		return false
	}
	s.mu.Lock()
	delete(s.attempts, ip)
	s.mu.Unlock()
	return true
}

func (s *identityStore) issue(a account) (string, bool) {
	now := time.Now()
	token := rand.Text()
	s.mu.Lock()
	defer s.mu.Unlock()
	for key, entry := range s.sessions {
		if !entry.expires.After(now) {
			delete(s.sessions, key)
		}
	}
	if len(s.sessions) >= 8192 {
		return "", false
	}
	s.sessions[sha256.Sum256([]byte(token))] = session{name: a.Name, credential: a.Hash, expires: now.Add(sessionLifetime)}
	return token, true
}

func cookieKey(r *http.Request) ([32]byte, bool) {
	cookies := r.CookiesNamed(sessionCookie)
	if len(cookies) != 1 || cookies[0].Value == "" {
		return [32]byte{}, false
	}
	return sha256.Sum256([]byte(cookies[0].Value)), true
}

func (s *identityStore) principal(r *http.Request) (account, error) {
	key, ok := cookieKey(r)
	if !ok {
		return account{}, os.ErrNotExist
	}
	s.mu.Lock()
	entry, ok := s.sessions[key]
	if ok && !time.Now().Before(entry.expires) {
		delete(s.sessions, key)
		ok = false
	}
	s.mu.Unlock()
	if !ok {
		return account{}, os.ErrNotExist
	}
	a, err := s.load(entry.name)
	if err == nil && entry.credential != a.Hash {
		err = os.ErrNotExist
	}
	return a, err
}

func (s *identityStore) revoke(r *http.Request) {
	if key, ok := cookieKey(r); ok {
		s.mu.Lock()
		delete(s.sessions, key)
		s.mu.Unlock()
	}
}
