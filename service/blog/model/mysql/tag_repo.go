package mysql

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	blogrepository "sitehub/service/blog/model"
	blogentity "sitehub/service/blog/model/entity"
)

// TagRepositoryImpl 标签仓储实现
type TagRepositoryImpl struct {
	db *gorm.DB
}

// NewTagRepositoryImpl 创建标签仓储实例
func NewTagRepositoryImpl(db *gorm.DB) blogrepository.TagRepository {
	return &TagRepositoryImpl{db: db}
}

// Create 创建标签
func (r *TagRepositoryImpl) Create(tag *blogentity.Tag) error {
	return r.db.Create(tag).Error
}

func (r *TagRepositoryImpl) GetAllTags() ([]*blogentity.Tag, error) {
	var tags []*blogentity.Tag
	err := r.db.Find(&tags).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get all tags: %w", err)
	}
	return tags, nil
}

// Update 更新标签
func (r *TagRepositoryImpl) Update(tag *blogentity.Tag) error {
	// 使用 Updates 更新非零字段，通过 ID 定位记录
	return r.db.Model(&blogentity.Tag{}).
		Where("id = ?", tag.ID).
		Select("Name", "Slug", "Description").
		Updates(tag).
		Error
}

// Delete 删除标签
func (r *TagRepositoryImpl) Delete(id uint) error {
	return r.db.Delete(&blogentity.Tag{}, id).Error
}

// GetByID 根据ID获取标签
func (r *TagRepositoryImpl) GetByID(id uint) (*blogentity.Tag, error) {
	var tag blogentity.Tag
	err := r.db.First(&tag, id).Error
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

// GetBySlug 根据slug获取标签
func (r *TagRepositoryImpl) GetBySlug(slug string) (*blogentity.Tag, error) {
	var tag blogentity.Tag
	err := r.db.Where("slug = ?", slug).First(&tag).Error
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

// List 分页获取标签列表
func (r *TagRepositoryImpl) List(page, pageSize int, name string) ([]*blogentity.Tag, int64, error) {
	var tags []*blogentity.Tag
	var total int64

	offset := (page - 1) * pageSize
	query := r.db.Model(&blogentity.Tag{})
	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("id ASC").Offset(offset).Limit(pageSize).Find(&tags).Error
	if err != nil {
		return nil, 0, err
	}

	// 为每个标签填充文章数量
	for _, tag := range tags {
		var count int64
		if err := r.db.Table("articles a").Joins("LEFT JOIN article_publications ap ON ap.article_id=a.id").Where("a.status=1 AND a.deleted_at IS NULL").Where("a.id IN (SELECT article_id FROM article_publication_tags WHERE tag_id=?) OR (ap.article_id IS NULL AND a.id IN (SELECT article_id FROM article_tag WHERE tag_id=?))", tag.ID, tag.ID).Count(&count).Error; err != nil {
			return nil, 0, err
		}
		tag.ArticleCount = count
	}

	return tags, total, err
}

// GetByArticleID 根据文章ID获取标签列表
func (r *TagRepositoryImpl) GetByArticleID(articleID uint) ([]*blogentity.Tag, error) {
	repo := &ArticleRepositoryImpl{db: r.db}
	row, err := repo.WorkflowGet(uint64(articleID), false)
	if err != nil {
		return nil, err
	}
	if row.Status != 1 || row.DeletedAt != nil {
		return nil, gorm.ErrRecordNotFound
	}
	if err = repo.PublicationModels([]*blogentity.Article{row}); err != nil {
		return nil, err
	}
	tags := []*blogentity.Tag{}
	for i := range row.Tags {
		tags = append(tags, &row.Tags[i])
	}
	return tags, nil
}
func (r *TagRepositoryImpl) GetTagArticles(tagID uint, page, pageSize int) ([]*blogentity.Article, int64, error) {
	repo := &ArticleRepositoryImpl{db: r.db}
	rows, total, err := repo.WorkflowList(page, pageSize, "", "latest", nil, 0, []uint{tagID}, false, true)
	if err != nil {
		return nil, 0, err
	}
	err = repo.PublicationModels(rows)
	return rows, total, err
}

func (r *TagRepositoryImpl) WithTransaction(fn func(blogrepository.TagRepository) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error { return fn(&TagRepositoryImpl{db: tx}) })
}
func (r *TagRepositoryImpl) LockByID(id uint) error {
	var row blogentity.Tag
	return r.db.Clauses(clause.Locking{Strength: "UPDATE"}).First(&row, id).Error
}
func (r *TagRepositoryImpl) ReferenceCount(id uint) (int64, error) {
	var a, b int64
	if err := r.db.Model(&blogentity.ArticleTag{}).Where("tag_id = ?", id).Count(&a).Error; err != nil {
		return 0, err
	}
	err := r.db.Model(&blogentity.ArticlePublicationTag{}).Where("tag_id = ?", id).Count(&b).Error
	return a + b, err
}
