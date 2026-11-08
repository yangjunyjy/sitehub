package mysql

import (
	"encoding/json"
	"errors"
	"sort"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	blogdomain "sitehub/service/blog/domain"
	blogrepository "sitehub/service/blog/model"
	blogentity "sitehub/service/blog/model/entity"
)

func (r *ArticleRepositoryImpl) WorkflowTransaction(fn func(blogrepository.ArticleRepository) error) error {
	return r.WithTransaction(fn)
}
func (r *ArticleRepositoryImpl) FindOperation(actor uint64, key string) (*blogentity.ArticleOperation, error) {
	var row blogentity.ArticleOperation
	err := r.db.Where("actor_id = ? AND request_id = ?", actor, key).First(&row).Error
	return &row, err
}
func (r *ArticleRepositoryImpl) StoreOperation(row *blogentity.ArticleOperation) error {
	return r.db.Create(row).Error
}
func (r *ArticleRepositoryImpl) WorkflowSave(row *blogentity.Article) error {
	if row.ID == 0 {
		return r.db.Omit(clause.Associations).Create(row).Error
	}
	return r.db.Model(row).Omit(clause.Associations).Select("*").Updates(row).Error
}
func (r *ArticleRepositoryImpl) WorkflowGet(id uint64, lock bool) (*blogentity.Article, error) {
	var row blogentity.Article
	q := r.db
	if lock {
		q = q.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	err := q.First(&row, id).Error
	if err != nil {
		return nil, err
	}
	if err = r.db.Model(&row).Association("Tags").Find(&row.Tags); err != nil {
		return nil, err
	}
	if row.CategoryID != 0 {
		var category blogentity.Category
		if err = r.db.First(&category, row.CategoryID).Error; err == nil {
			row.Category = &category
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}
	return &row, nil
}
func (r *ArticleRepositoryImpl) WorkflowList(page, size int, title, order string, status *int, category uint64, tags []uint, trash, public bool) ([]*blogentity.Article, int64, error) {
	q := r.db.Model(&blogentity.Article{})
	if trash {
		q = q.Where("articles.deleted_at IS NOT NULL")
	} else {
		q = q.Where("articles.deleted_at IS NULL")
	}
	if public {
		q = q.Joins("JOIN article_publications ap ON ap.article_id = articles.id").Where("articles.status = 1")
	}
	if status != nil && !public {
		q = q.Where("articles.status = ?", *status)
	}
	if title != "" {
		if public {
			pattern := "%" + title + "%"
			q = q.Where("ap.title LIKE ? OR ap.content LIKE ?", pattern, pattern)
		} else {
			q = q.Where("articles.title LIKE ?", "%"+title+"%")
		}
	}
	if category != 0 {
		if public {
			q = q.Where("ap.category_id = ?", category)
		} else {
			q = q.Where("articles.category_id = ?", category)
		}
	}
	if len(tags) > 0 {
		if public {
			q = q.Where("articles.id IN (SELECT article_id FROM article_publication_tags WHERE tag_id IN ?)", tags)
		} else {
			q = q.Where("articles.id IN (SELECT article_id FROM article_tag WHERE tag_id IN ?)", tags)
		}
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	ordering := "articles.updated_at DESC, articles.id DESC"
	if public {
		ordering = "articles.is_top DESC, articles.published_at DESC, articles.id DESC"
	}
	if order == "hot" {
		ordering = "articles.views DESC, articles.id DESC"
	}
	var rows []*blogentity.Article
	err := q.Select("articles.*").Preload("Category").Preload("Tags").Order(ordering).Offset((page - 1) * size).Limit(size).Find(&rows).Error
	return rows, total, err
}

// Parent locks use the same ordering as category/tag deletion. Every writer validates references here.
func (r *ArticleRepositoryImpl) ValidateReferences(category uint64, tags []uint, assets []string, actor uint64, articleID uint64) error {
	if category != 0 {
		var row blogentity.Category
		if err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).First(&row, category).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return blogdomain.ErrArticleReference
			}
			return err
		}
	}
	ids := append([]uint(nil), tags...)
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	if len(ids) > 0 {
		var rows []blogentity.Tag
		if err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id IN ?", ids).Order("id").Find(&rows).Error; err != nil {
			return err
		}
		if len(rows) != len(ids) {
			return blogdomain.ErrArticleReference
		}
	}
	sort.Strings(assets)
	for _, id := range assets {
		var asset blogentity.ArticleAsset
		if err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND deleted_at IS NULL", id).First(&asset).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return blogdomain.ErrArticleReference
			}
			return err
		}
		if asset.Location == "" {
			return blogdomain.ErrArticleReference
		}
		if asset.OwnerID != actor {
			var count int64
			if err := r.db.Model(&blogentity.ArticleAssetReference{}).Where("asset_id = ? AND article_id = ?", id, articleID).Count(&count).Error; err != nil {
				return err
			}
			if count == 0 {
				return blogdomain.ErrArticleReference
			}
		}
	}
	return nil
}
func (r *ArticleRepositoryImpl) SaveRevision(row *blogentity.ArticleRevision) error {
	return r.db.Create(row).Error
}
func (r *ArticleRepositoryImpl) Revisions(id uint64) ([]blogentity.ArticleRevision, error) {
	rows := []blogentity.ArticleRevision{}
	err := r.db.Where("article_id = ?", id).Order("version DESC").Limit(100).Find(&rows).Error
	return rows, err
}
func (r *ArticleRepositoryImpl) Revision(id, revisionID uint64) (*blogentity.ArticleRevision, error) {
	var row blogentity.ArticleRevision
	err := r.db.Where("article_id = ? AND id = ?", id, revisionID).First(&row).Error
	return &row, err
}
func (r *ArticleRepositoryImpl) Publication(id uint64) (*blogentity.ArticlePublication, error) {
	var row blogentity.ArticlePublication
	err := r.db.First(&row, "article_id = ?", id).Error
	return &row, err
}
func (r *ArticleRepositoryImpl) Publish(row *blogentity.ArticlePublication, tags []uint) error {
	if err := r.db.Clauses(clause.OnConflict{UpdateAll: true}).Create(row).Error; err != nil {
		return err
	}
	if err := r.db.Where("article_id = ?", row.ArticleID).Delete(&blogentity.ArticlePublicationTag{}).Error; err != nil {
		return err
	}
	for _, tag := range tags {
		if err := r.db.Create(&blogentity.ArticlePublicationTag{ArticleID: row.ArticleID, TagID: uint64(tag)}).Error; err != nil {
			return err
		}
	}
	return nil
}
func (r *ArticleRepositoryImpl) SaveAssetReferences(id, version uint64, assets []string) error {
	for _, asset := range assets {
		if err := r.db.Create(&blogentity.ArticleAssetReference{ArticleID: id, Version: version, AssetID: asset}).Error; err != nil {
			return err
		}
	}
	return nil
}
func (r *ArticleRepositoryImpl) Options() ([]blogentity.Category, []blogentity.Tag, error) {
	categories, tags := []blogentity.Category{}, []blogentity.Tag{}
	if err := r.db.Order("name,id").Find(&categories).Error; err != nil {
		return nil, nil, err
	}
	err := r.db.Order("name,id").Find(&tags).Error
	return categories, tags, err
}

func (r *ArticleRepositoryImpl) RevisionByVersion(id, version uint64) (*blogentity.ArticleRevision, error) {
	var row blogentity.ArticleRevision
	err := r.db.Where("article_id = ? AND version = ?", id, version).First(&row).Error
	return &row, err
}

// Public category/tag views always read the immutable publication snapshot.
func (r *ArticleRepositoryImpl) PublicationModels(rows []*blogentity.Article) error {
	for _, row := range rows {
		publication, err := r.Publication(row.ID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return blogdomain.ErrArticleNotFound
		}
		if err != nil {
			return err
		}
		var v blogdomain.ArticleView
		if err = json.Unmarshal([]byte(publication.Payload), &v); err != nil {
			return err
		}
		row.Title = v.Title
		row.Summary = v.Summary
		row.Content = v.Content
		row.Slug = v.Slug
		row.Cover = v.Cover
		row.ContentFormat = v.ContentFormat
		row.CategoryID = v.CategoryID
		row.Tags = []blogentity.Tag{}
		for _, tag := range v.Tags {
			row.Tags = append(row.Tags, blogentity.Tag{ID: tag.ID, Name: tag.Name})
		}
		row.Category = nil
		if v.Category != nil {
			row.Category = &blogentity.Category{ID: uint(v.Category.ID), Name: v.Category.Name}
		}
		row.Author = &blogentity.User{ID: uint(v.Author.ID), Username: v.Author.Name}
	}
	return nil
}
