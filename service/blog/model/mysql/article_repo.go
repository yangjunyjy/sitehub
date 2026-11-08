package mysql

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	blogrepository "sitehub/service/blog/model"
	blogentity "sitehub/service/blog/model/entity"
)

// ArticleRepositoryImpl 文章仓储实现
type ArticleRepositoryImpl struct {
	db *gorm.DB
}

// NewArticleRepositoryImpl 创建文章仓储实例
func NewArticleRepositoryImpl(db *gorm.DB) blogrepository.ArticleRepository {
	return &ArticleRepositoryImpl{db: db}
}

// Create 创建文章
func (r *ArticleRepositoryImpl) Create(article *blogentity.Article) error {
	return r.db.Create(article).Error
}

// Update 更新文章
func (r *ArticleRepositoryImpl) Update(article *blogentity.Article) error {
	return r.db.Model(&blogentity.Article{}).Where("id = ?", article.ID).
		Select("title", "slug", "summary", "content", "cover", "category_id", "status", "is_top", "sort_order", "published_at", "scheduled_at", "updated_at").Updates(article).Error
}

// Delete 删除文章
func (r *ArticleRepositoryImpl) Delete(id uint) error {
	return r.db.Delete(&blogentity.Article{}, id).Error

}

// GetByID 根据ID获取文章
func (r *ArticleRepositoryImpl) GetByID(id uint) (*blogentity.Article, error) {
	var article blogentity.Article
	err := r.db.Preload("Category").
		Preload("Tags").
		First(&article, id).Error
	if err != nil {
		return nil, err
	}
	return &article, nil
}

// List 分页获取文章列表
func (r *ArticleRepositoryImpl) List(page,
	pageSize int,
	title string,
	status *int,
	categoryID *uint,
	tags []uint) ([]*blogentity.Article, int64, error) {
	var articles []*blogentity.Article
	var total int64

	offset := (page - 1) * pageSize
	query := r.db.Model(&blogentity.Article{}).
		Select("articles.id, articles.title, articles.slug, articles.summary, articles.cover, articles.category_id, articles.author_id, articles.views, articles.likes, articles.favorites, articles.comment_count, articles.status, articles.is_top, articles.sort_order, articles.published_at, articles.created_at, articles.updated_at").
		Preload("Category").
		Preload("Tags")
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	if categoryID != nil {
		query = query.Where("category_id = ?", *categoryID)
	}
	if title != "" {
		query = query.Where("title LIKE ?", "%"+title+"%")
	}
	if len(tags) > 0 {
		// A semi-join keeps each article once when several selected tags match it.
		// It also avoids ambiguous created_at columns in the order clause.
		query = query.Where("articles.id IN (?)", r.db.Model(&blogentity.ArticleTag{}).Select("article_id").Where("tag_id IN ?", tags))
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Offset(offset).Limit(pageSize).Order("articles.is_top DESC, articles.sort_order ASC, articles.created_at DESC").Find(&articles).Error

	return articles, total, err
}

// GetByCategoryID 根据分类ID获取文章列表
func (r *ArticleRepositoryImpl) GetByCategoryID(categoryID uint, page, pageSize int) ([]*blogentity.Article, int64, error) {
	var articles []*blogentity.Article
	var total int64

	offset := (page - 1) * pageSize
	query := r.db.Model(&blogentity.Article{}).
		Select("id, title, slug, summary, cover, category_id, author_id, views, likes, favorites, comment_count, status, is_top, sort_order, published_at, created_at, updated_at").
		Where("category_id = ?", categoryID).
		Preload("Category").
		Preload("Tags")
	query.Count(&total)
	err := query.Offset(offset).Limit(pageSize).Order("is_top DESC, sort_order ASC, created_at DESC").Find(&articles).Error

	return articles, total, err
}

// GetByTagID 根据标签ID获取文章列表
func (r *ArticleRepositoryImpl) GetByTagID(tagID uint, page, pageSize int) ([]*blogentity.Article, int64, error) {
	var articles []*blogentity.Article
	var total int64

	offset := (page - 1) * pageSize
	query := r.db.Model(&blogentity.Article{}).
		Select("articles.id, articles.title, articles.slug, articles.summary, articles.cover, articles.category_id, articles.author_id, articles.views, articles.likes, articles.favorites, articles.comment_count, articles.status, articles.is_top, articles.sort_order, articles.published_at, articles.created_at, articles.updated_at").
		Joins("JOIN article_tag ON articles.id = article_tag.article_id").
		Where("article_tag.tag_id = ?", tagID).
		Preload("Category").
		Preload("Tags")
	query.Count(&total)
	err := query.Offset(offset).Limit(pageSize).Order("is_top DESC, sort_order ASC, created_at DESC").Find(&articles).Error

	return articles, total, err
}

// Search 搜索文章
func (r *ArticleRepositoryImpl) Search(keyword string, page, pageSize int) ([]*blogentity.Article, int64, error) {
	var articles []*blogentity.Article
	var total int64

	offset := (page - 1) * pageSize
	query := r.db.Model(&blogentity.Article{}).
		Select("id, title, slug, summary, cover, category_id, author_id, views, likes, favorites, comment_count, status, is_top, sort_order, published_at, created_at, updated_at").
		Where("status = ? AND (title LIKE ? OR content LIKE ?)", 1, "%"+keyword+"%", "%"+keyword+"%").
		Preload("Category").
		Preload("Tags")
	query.Count(&total)
	err := query.Offset(offset).Limit(pageSize).Order("is_top DESC, sort_order ASC, created_at DESC").Find(&articles).Error

	return articles, total, err
}

// IncrementViewCount 增加浏览次数
func (r *ArticleRepositoryImpl) IncrementViewCount(id uint) error {
	return r.db.Model(&blogentity.Article{}).Where("id = ?", id).UpdateColumn("views", gorm.Expr("views + 1")).Error
}

// GetHotArticles 获取热门文章
func (r *ArticleRepositoryImpl) GetHotArticles(limit int) ([]*blogentity.Article, error) {
	var articles []*blogentity.Article
	err := r.db.Select("id, title, slug, summary, cover, category_id, author_id, views, likes, favorites, comment_count, status, is_top, sort_order, published_at, created_at, updated_at").
		Preload("Category").
		Preload("Tags").
		Where("status = ?", 1).Order("views DESC").Limit(limit).Find(&articles).Error
	return articles, err
}

// GetRecentArticles 获取最新文章
func (r *ArticleRepositoryImpl) GetRecentArticles(limit int) ([]*blogentity.Article, error) {
	var articles []*blogentity.Article
	err := r.db.Select("id, title, slug, summary, cover, category_id, author_id, views, likes, favorites, comment_count, status, is_top, sort_order, published_at, created_at, updated_at").
		Preload("Category").
		Preload("Tags").
		Where("status = ?", 1).Order("is_top DESC, created_at DESC").Limit(limit).Find(&articles).Error
	return articles, err
}

// SaveArticleTags 保存文章标签关联
func (r *ArticleRepositoryImpl) SaveArticleTags(articleID uint, tagIDs []uint) error {
	if len(tagIDs) == 0 {
		return nil
	}

	// 批量插入标签关联
	var articleTags []blogentity.ArticleTag
	for _, tagID := range tagIDs {
		articleTags = append(articleTags, blogentity.ArticleTag{
			ArticleID: uint64(articleID),
			TagID:     uint64(tagID),
		})
	}

	return r.db.Create(&articleTags).Error
}

// DeleteArticleTags 删除文章的所有标签关联
func (r *ArticleRepositoryImpl) DeleteArticleTags(articleID uint) error {
	return r.db.Where("article_id = ?", articleID).Delete(&blogentity.ArticleTag{}).Error
}

// GetScheduledArticles 获取待定时发布的文章
func (r *ArticleRepositoryImpl) GetScheduledArticles(after time.Time) ([]*blogentity.Article, error) {
	var articles []*blogentity.Article
	err := r.db.Where("status = 2 AND scheduled_at > ? AND deleted_at IS NULL", after).
		Order("scheduled_at ASC").
		Find(&articles).Error
	return articles, err
}

// ToggleTop 切换置顶状态（使用 UpdateColumn 避免 GORM 零值跳过问题）
func (r *ArticleRepositoryImpl) ToggleTop(id uint, isTop bool) error {
	result := r.db.Model(&blogentity.Article{}).Where("id = ?", id).UpdateColumn("is_top", isTop)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *ArticleRepositoryImpl) WithTransaction(fn func(blogrepository.ArticleRepository) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error { return fn(&ArticleRepositoryImpl{db: tx}) })
}

func (r *ArticleRepositoryImpl) LockByID(id uint) (*blogentity.Article, error) {
	var article blogentity.Article
	err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).First(&article, id).Error
	return &article, err
}

func (r *ArticleRepositoryImpl) DeleteArticleViews(id uint) error {
	return r.db.Where("article_id = ?", id).Delete(&blogentity.ArticleView{}).Error
}

// 与删除路径锁定同一文章；只记录已发布文章，避免产生孤立浏览记录。
func (r *ArticleRepositoryImpl) RecordView(articleID, userID uint, ipAddress, userAgent string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		repo := &ArticleRepositoryImpl{db: tx}
		article, err := repo.LockByID(articleID)
		if err != nil {
			return err
		}
		if article.Status != 1 || article.DeletedAt != nil {
			return gorm.ErrRecordNotFound
		}
		view := &blogentity.ArticleView{ArticleID: uint64(articleID), IPAddress: ipAddress, UserAgent: userAgent}
		if err := tx.Create(view).Error; err != nil {
			return err
		}
		return tx.Model(&blogentity.Article{}).Where("id = ?", articleID).UpdateColumn("views", gorm.Expr("views + 1")).Error
	})
}

func (r *ArticleRepositoryImpl) WithContext(ctx context.Context) blogrepository.ArticleRepository {
	return &ArticleRepositoryImpl{db: r.db.WithContext(ctx)}
}
