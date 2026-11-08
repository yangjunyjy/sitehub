package repository

import (
	"context"
	"time"

	blogentity "sitehub/service/blog/model/entity"
)

// ArticleRepository 文章仓储接口
type ArticleRepository interface {
	WithContext(context.Context) ArticleRepository
	AssetCreate(*blogentity.ArticleAsset) error
	AssetGet(string) (*blogentity.ArticleAsset, error)
	AssetUpdate(*blogentity.ArticleAsset) error
	AssetPublished(string) (bool, error)
	AssetCleanupCandidates(time.Time) ([]blogentity.ArticleAsset, error)
	AssetRetire(string) error
	AssetForget(string) error

	WorkflowTransaction(func(ArticleRepository) error) error
	FindOperation(actor uint64, key string) (*blogentity.ArticleOperation, error)
	StoreOperation(operation *blogentity.ArticleOperation) error
	WorkflowSave(article *blogentity.Article) error
	WorkflowGet(id uint64, lock bool) (*blogentity.Article, error)
	WorkflowList(page, size int, title, sort string, status *int, category uint64, tags []uint, trash, public bool) ([]*blogentity.Article, int64, error)
	ValidateReferences(category uint64, tags []uint, assets []string, actor uint64, articleID uint64) error
	SaveRevision(revision *blogentity.ArticleRevision) error
	RevisionByVersion(id, version uint64) (*blogentity.ArticleRevision, error)
	Revisions(id uint64) ([]blogentity.ArticleRevision, error)
	Revision(id, revisionID uint64) (*blogentity.ArticleRevision, error)
	Publication(id uint64) (*blogentity.ArticlePublication, error)
	Publish(publication *blogentity.ArticlePublication, tags []uint) error
	SaveAssetReferences(id, version uint64, assets []string) error
	Options() ([]blogentity.Category, []blogentity.Tag, error)

	WithTransaction(func(ArticleRepository) error) error
	LockByID(id uint) (*blogentity.Article, error)
	DeleteArticleViews(id uint) error
	Create(article *blogentity.Article) error
	Update(article *blogentity.Article) error
	Delete(id uint) error
	GetByID(id uint) (*blogentity.Article, error)
	List(page, pageSize int, title string, status *int, categoryID *uint, tags []uint) ([]*blogentity.Article, int64, error)
	Search(keyword string, page, pageSize int) ([]*blogentity.Article, int64, error)
	IncrementViewCount(id uint) error
	GetHotArticles(limit int) ([]*blogentity.Article, error)
	GetRecentArticles(limit int) ([]*blogentity.Article, error)

	// 标签关联操作
	SaveArticleTags(articleID uint, tagIDs []uint) error
	DeleteArticleTags(articleID uint) error

	// 定时发布相关
	GetScheduledArticles(after time.Time) ([]*blogentity.Article, error) // 获取待定时发布的文章

	ToggleTop(id uint, isTop bool) error                                  // 切换置顶状态（使用 UpdateColumn 避免 GORM 零值跳过）
	RecordView(articleID, userID uint, ipAddress, userAgent string) error // 记录浏览记录到 article_view 表
}
