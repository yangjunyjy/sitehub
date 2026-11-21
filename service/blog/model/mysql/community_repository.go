package mysql

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"sitehub/service/blog/model/entity"
)

type CommunityRepository struct{ db *gorm.DB }

func (repo *CommunityRepository) SaveHome(ctx context.Context, animation bool, verse string) error {
	profile := entity.AuthorProfile{ID: 1, HomeAnimation: &animation, HomeVerse: verse}
	return repo.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "id"}}, DoUpdates: clause.AssignmentColumns([]string{"home_animation", "home_verse"}),
	}).Create(&profile).Error
}

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
func (repo *CommunityRepository) PublishedMessages(ctx context.Context, articleID uint64, offset, limit int) ([]entity.GuestMessage, int64, error) {
	var messages []entity.GuestMessage
	var total int64
	query := repo.db.WithContext(ctx).Model(&entity.GuestMessage{}).Where("status = ? AND article_id = ?", "approved", articleID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("id DESC").Offset(offset).Limit(limit).Find(&messages).Error
	return messages, total, err
}
func (repo *CommunityRepository) Submit(ctx context.Context, message *entity.GuestMessage) error {
	db := repo.db.WithContext(ctx)
	if message.ArticleID == 0 {
		return db.Create(message).Error
	}
	return db.Transaction(func(tx *gorm.DB) error {
		if err := lockPublished(tx, message.ArticleID); err != nil {
			return err
		}
		return tx.Create(message).Error
	})
}
