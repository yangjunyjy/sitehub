package mysql

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"sitehub/service/blog/domain"
	"sitehub/service/blog/model/entity"
)

type ReviewItem struct {
	ID        uint64    `json:"id,string"`
	ArticleID uint64    `json:"articleId,string"`
	Name      string    `json:"name"`
	URL       string    `json:"url"`
	Content   string    `json:"content"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

func reviewQuery(db *gorm.DB, kind string) *gorm.DB {
	if kind == "friends" {
		return db.Model(&entity.FriendLink{})
	}
	query := db.Model(&entity.GuestMessage{})
	if kind == "comments" {
		return query.Where("article_id > 0")
	}
	return query.Where("article_id = 0")
}

func (repo *CommunityRepository) ReviewQueue(ctx context.Context, kind, state string, page int) ([]ReviewItem, int64, error) {
	query := reviewQuery(repo.db.WithContext(ctx), kind).Where("status = ?", state)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	fields := "id, article_id, nickname AS name, content, status, created_at"
	if kind == "friends" {
		fields = "id, name, url, description AS content, status, created_at"
	}
	items := []ReviewItem{}
	err := query.Select(fields).Order("id DESC").Offset((page - 1) * 20).Limit(20).Scan(&items).Error
	return items, total, err
}

func (repo *CommunityRepository) Review(ctx context.Context, kind string, id uint64, state string) error {
	return repo.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		query := reviewQuery(tx, kind).Where("id = ?", id)
		fields := "id, status"
		if kind == "comments" {
			fields += ", article_id"
		}
		var current ReviewItem
		if err := query.Session(&gorm.Session{}).Select(fields).Clauses(clause.Locking{Strength: "UPDATE"}).Take(&current).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domain.ErrCommunityItemNotFound
			}
			return err
		}
		if current.Status == state {
			return nil
		}
		if err := query.Update("status", state).Error; err != nil {
			return err
		}
		if kind != "comments" {
			return nil
		}
		articles := tx.Model(&entity.Article{}).Where("id = ?", current.ArticleID)
		if state == "approved" {
			return articles.UpdateColumn("comment_count", gorm.Expr("comment_count + 1")).Error
		}
		if current.Status == "approved" {
			return articles.Where("comment_count > 0").UpdateColumn("comment_count", gorm.Expr("comment_count - 1")).Error
		}
		return nil
	})
}
