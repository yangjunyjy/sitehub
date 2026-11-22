package domain

import (
	"errors"
	"time"
)

var ErrMemberMissing = errors.New("product member not found")
var ErrMemberCredentials = errors.New("账号或密码不正确")
var ErrMemberSession = errors.New("登录已过期，请重新登录")
var ErrMemberRateLimit = errors.New("操作过于频繁，请稍后重试")

// Member 与 sys_user 和历史文章作者 user 完全独立，不携带后台角色。
type Member struct {
	ID           uint64
	Email        string  `gorm:"size:254;uniqueIndex"`
	Username     *string `gorm:"size:40;uniqueIndex"`
	Nickname     string  `gorm:"size:40"`
	Avatar       string  `gorm:"size:1024"`
	PasswordHash string
	CreatedAt    time.Time
}

func (Member) TableName() string { return "product_members" }

type MemberVerification struct {
	Email     string `gorm:"primaryKey"`
	CodeHash  string
	ExpiresAt time.Time
	Attempts  int
}

func (MemberVerification) TableName() string { return "product_member_verifications" }

type MemberSession struct {
	TokenHash string `gorm:"primaryKey"`
	MemberID  uint64
	ExpiresAt time.Time
}

func (MemberSession) TableName() string { return "product_member_sessions" }

type RegisterMemberInput struct{ Email, Password, Code, Username, Nickname string }
type UpdateMemberProfileInput struct{ Username, Nickname, Avatar string }
type ResetMemberPasswordInput struct{ Email, Password, Code string }
type LoginMemberInput struct{ Account, Email, Password string }

type MemberProfile struct {
	Username  string    `json:"username"`
	Nickname  string    `json:"nickname"`
	Avatar    string    `json:"avatar"`
	ID        uint64    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
}

func (m Member) Profile() MemberProfile {
	profile := MemberProfile{ID: m.ID, Email: m.Email, CreatedAt: m.CreatedAt, Nickname: m.Nickname, Avatar: m.Avatar}
	if m.Username != nil {
		profile.Username = *m.Username
	}
	return profile
}
