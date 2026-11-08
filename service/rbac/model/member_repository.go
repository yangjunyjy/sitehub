package repository

import (
	"context"
	"time"

	rbacdomain "sitehub/service/rbac/domain"
)

type MemberRepository interface {
	WithContext(context.Context) MemberRepository
	Transaction(func(MemberRepository) error) error
	MemberByEmail(string) (rbacdomain.Member, error)
	CreateMember(*rbacdomain.Member) error
	Verification(string, bool) (rbacdomain.MemberVerification, error)
	PutVerification(rbacdomain.MemberVerification) error
	DeleteVerification(string, string) error
	CreateSession(rbacdomain.MemberSession) error
	SessionMember(string, time.Time) (rbacdomain.Member, error)
	DeleteSession(string) error
	DeleteExpiredSessions(uint64, time.Time) error
}
