package mysql

import (
	"time"

	"gorm.io/gorm/clause"

	blogdomain "sitehub/service/blog/domain"
	blogentity "sitehub/service/blog/model/entity"
)

func (r *ArticleRepositoryImpl) AssetCreate(a *blogentity.ArticleAsset) error {
	return r.db.Create(a).Error
}
func (r *ArticleRepositoryImpl) AssetGet(id string) (*blogentity.ArticleAsset, error) {
	var a blogentity.ArticleAsset
	err := r.db.First(&a, "id = ?", id).Error
	return &a, err
}
func (r *ArticleRepositoryImpl) AssetUpdate(a *blogentity.ArticleAsset) error {
	return r.db.Model(a).Select("*").Updates(a).Error
}
func (r *ArticleRepositoryImpl) AssetPublished(id string) (bool, error) {
	var n int64
	err := r.db.Table("article_asset_references ar").Joins("JOIN articles a ON a.id = ar.article_id AND a.published_version = ar.version").Where("ar.asset_id = ? AND a.status = 1 AND a.deleted_at IS NULL", id).Count(&n).Error
	return n > 0, err
}
func (r *ArticleRepositoryImpl) AssetCleanupCandidates(before time.Time) ([]blogentity.ArticleAsset, error) {
	rows := []blogentity.ArticleAsset{}
	err := r.db.Where("(deleted_at IS NOT NULL OR created_at < ?) AND id NOT IN (SELECT asset_id FROM article_asset_references)", before).Limit(100).Find(&rows).Error
	return rows, err
}
func (r *ArticleRepositoryImpl) AssetRetire(id string) error {
	var a blogentity.ArticleAsset
	if err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).First(&a, "id = ?", id).Error; err != nil {
		return err
	}
	var count int64
	if err := r.db.Model(&blogentity.ArticleAssetReference{}).Where("asset_id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count != 0 {
		return blogdomain.ErrArticleState
	}
	now := time.Now()
	return r.db.Model(&a).Update("deleted_at", now).Error
}
func (r *ArticleRepositoryImpl) AssetForget(id string) error {
	return r.db.Where("id = ? AND deleted_at IS NOT NULL", id).Delete(&blogentity.ArticleAsset{}).Error
}
