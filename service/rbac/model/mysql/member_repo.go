package mysql

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	rbacdomain "sitehub/service/rbac/domain"
	rbacrepository "sitehub/service/rbac/model"
)

type memberRepository struct{ db *gorm.DB }

func NewMemberRepository(db *gorm.DB) rbacrepository.MemberRepository { return &memberRepository{db} }
func (r *memberRepository) WithContext(ctx context.Context) rbacrepository.MemberRepository {
	return &memberRepository{r.db.WithContext(ctx)}
}
func (r *memberRepository) Transaction(fn func(rbacrepository.MemberRepository) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error { return fn(&memberRepository{tx}) })
}
func memberLookupError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return rbacdomain.ErrMemberMissing
	}
	return err
}
func (r *memberRepository) MemberByUsername(username string) (m rbacdomain.Member, err error) {
	err = memberLookupError(r.db.Where("username = ?", username).Take(&m).Error)
	return
}
func (r *memberRepository) MemberByEmail(email string) (m rbacdomain.Member, err error) {
	err = memberLookupError(r.db.Where("email = ?", email).Take(&m).Error)
	return
}
func (r *memberRepository) CreateMember(m *rbacdomain.Member) error { return r.db.Create(m).Error }
func (r *memberRepository) Verification(email string, lock bool) (v rbacdomain.MemberVerification, err error) {
	q := r.db.Where("email = ?", email)
	if lock {
		q = q.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	err = memberLookupError(q.Take(&v).Error)
	return
}
func (r *memberRepository) PutVerification(v rbacdomain.MemberVerification) error {
	return r.db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "email"}}, DoUpdates: clause.AssignmentColumns([]string{"code_hash", "expires_at", "attempts"})}).Create(&v).Error
}
func (r *memberRepository) DeleteVerification(email, hash string) error {
	return r.db.Where("email = ? AND code_hash = ?", email, hash).Delete(&rbacdomain.MemberVerification{}).Error
}
func (r *memberRepository) CreateSession(v rbacdomain.MemberSession) error {
	return r.db.Create(&v).Error
}
func (r *memberRepository) SessionMember(hash string, now time.Time) (m rbacdomain.Member, err error) {
	err = memberLookupError(r.db.Table("product_members AS m").Select("m.*").Joins("JOIN product_member_sessions s ON s.member_id = m.id").Where("s.token_hash = ? AND s.expires_at > ?", hash, now).Take(&m).Error)
	return
}
func (r *memberRepository) DeleteSession(hash string) error {
	return r.db.Where("token_hash = ?", hash).Delete(&rbacdomain.MemberSession{}).Error
}
func (r *memberRepository) DeleteExpiredSessions(id uint64, now time.Time) error {
	return r.db.Where("member_id = ? AND expires_at <= ?", id, now).Delete(&rbacdomain.MemberSession{}).Error
}

func (r *memberRepository) LockMember(id uint64) (m rbacdomain.Member, err error) {
	err = memberLookupError(r.db.Clauses(clause.Locking{Strength: "UPDATE"}).First(&m, id).Error)
	return
}
func (r *memberRepository) SaveProfile(id uint64, profile rbacdomain.UpdateMemberProfileInput) error {
	return r.db.Model(&rbacdomain.Member{}).Where("id = ?", id).Updates(map[string]any{"username": profile.Username, "nickname": profile.Nickname, "avatar": profile.Avatar}).Error
}

// 调用方在同一事务持有会员行锁，避免重置后留下用旧密码签发的会话。
func (r *memberRepository) ResetPassword(id uint64, hash string) error {
	if err := r.db.Model(&rbacdomain.Member{}).Where("id = ?", id).Update("password_hash", hash).Error; err != nil {
		return err
	}
	return r.db.Where("member_id = ?", id).Delete(&rbacdomain.MemberSession{}).Error
}
