package mysql

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"sitehub/service/blog/model/entity"
)

type CommunityRepository struct{ db *gorm.DB }

func NewCommunityRepository(db *gorm.DB) *CommunityRepository { return &CommunityRepository{db: db} }
func (repo *CommunityRepository) Profile(ctx context.Context) (entity.AuthorProfile, error) {
	var profile entity.AuthorProfile
	err := repo.db.WithContext(ctx).First(&profile, 1).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return profile, err
}
func (repo *CommunityRepository) PublishedFriends(ctx context.Context) ([]entity.FriendLink, error) {
	var links []entity.FriendLink
	err := repo.db.WithContext(ctx).Where("status = ?", "approved").Order("id DESC").Find(&links).Error
	return links, err
}
func (repo *CommunityRepository) Apply(ctx context.Context, link *entity.FriendLink) error {
	return repo.db.WithContext(ctx).Create(link).Error
}
func (repo *CommunityRepository) PublishedMessages(ctx context.Context, offset, limit int) ([]entity.GuestMessage, error) {
	var messages []entity.GuestMessage
	err := repo.db.WithContext(ctx).Where("status = ?", "approved").Order("id DESC").Offset(offset).Limit(limit).Find(&messages).Error
	return messages, err
}
func (repo *CommunityRepository) Submit(ctx context.Context, message *entity.GuestMessage) error {
	return repo.db.WithContext(ctx).Create(message).Error
}
