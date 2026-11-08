package domain

import (
	"errors"
	"time"
)

var ErrMemberMissing = errors.New("product member not found")
var ErrMemberCredentials = errors.New("邮箱或密码不正确")
var ErrMemberSession = errors.New("登录已过期，请重新登录")
var ErrMemberRateLimit = errors.New("操作过于频繁，请稍后重试")

// Member 与 sys_user 和历史文章作者 user 完全独立，不携带后台角色。
type Member struct {
	ID           uint64
	Email        string `gorm:"size:254;uniqueIndex"`
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

type RegisterMemberInput struct{ Email, Password, Code string }
type LoginMemberInput struct{ Email, Password string }

type MemberProfile struct {
	ID        uint64    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
}

func (m Member) Profile() MemberProfile { return MemberProfile{m.ID, m.Email, m.CreatedAt} }
