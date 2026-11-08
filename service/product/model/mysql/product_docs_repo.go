package mysql

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	productdomain "sitehub/service/product/domain"
	productrepository "sitehub/service/product/model"
)

type productDocsRepository struct{ db *gorm.DB }

func NewProductDocsRepository(db *gorm.DB) productrepository.ProductDocsRepository {
	return &productDocsRepository{db}
}
func (r *productDocsRepository) Transaction(fn func(productrepository.ProductDocsRepository) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error { return fn(&productDocsRepository{tx}) })
}
func (r *productDocsRepository) Site(lock bool) (v productdomain.DocSite, err error) {
	q := r.db.Table("product_doc_sites").Where("id = 1")
	if lock {
		q = q.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	err = q.Take(&v).Error
	return
}
func (r *productDocsRepository) Versions() (v []productdomain.DocVersion, err error) {
	err = r.db.Table("product_doc_versions").Order("created_at, id").Find(&v).Error
	return
}
func (r *productDocsRepository) Modes() (v []productdomain.DocMode, err error) {
	err = r.db.Table("product_doc_modes").Order("sort_order, id").Find(&v).Error
	return
}
func (r *productDocsRepository) PutMode(v productdomain.DocMode) error {
	return r.db.Table("product_doc_modes").Save(&v).Error
}
func (r *productDocsRepository) DeleteMode(id string) error {
	result := r.db.Table("product_doc_modes").Where("id = ?", id).Delete(&productdomain.DocMode{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return productdomain.ErrDocumentNotFound
	}
	return nil
}
func (r *productDocsRepository) Groups(version string) (v []productdomain.DocGroup, err error) {
	q := r.db.Table("product_doc_groups")
	if version != "" {
		q = q.Where("version_id = ?", version)
	}
	err = q.Order("sort_order, id").Find(&v).Error
	return
}
func (r *productDocsRepository) Documents(version string, body bool) (v []productdomain.ProductDocument, err error) {
	q := r.db.Table("product_documents")
	if version != "" {
		q = q.Where("version_id = ?", version)
	}
	if !body {
		q = q.Omit("content_markdown")
	}
	err = q.Order("sort_order, id").Find(&v).Error
	return
}
func (r *productDocsRepository) Document(version, key string) (v productdomain.ProductDocument, err error) {
	// 管理操作与正文请求按稳定 ID 定位，不能因另一篇文档使用相同 slug 而命中错误记录。
	err = r.db.Table("product_documents").Where("version_id = ? AND doc_key = ?", version, key).Take(&v).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = productdomain.ErrDocumentNotFound
	}
	return
}
func (r *productDocsRepository) DocumentBySlug(version, slug string) (v productdomain.ProductDocument, err error) {
	err = r.db.Table("product_documents").Where("version_id = ? AND slug = ?", version, slug).Take(&v).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = productdomain.ErrDocumentNotFound
	}
	return
}
func (r *productDocsRepository) History(id uint64) (v []productdomain.DocumentRevision, err error) {
	err = r.db.Table("product_document_revisions").Omit("content_markdown").Where("document_id = ?", id).Order("revision DESC").Limit(100).Find(&v).Error
	return
}
func (r *productDocsRepository) Historical(id, history uint64) (v productdomain.DocumentRevision, err error) {
	err = r.db.Table("product_document_revisions").Where("document_id = ? AND id = ?", id, history).Take(&v).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = productdomain.ErrDocumentNotFound
	}
	return
}
func (r *productDocsRepository) PutSite(v productdomain.DocSite) error {
	return r.db.Table("product_doc_sites").Save(&v).Error
}
func (r *productDocsRepository) PutVersion(v productdomain.DocVersion) error {
	if v.CreatedAt.IsZero() {
		v.CreatedAt = time.Now()
	}
	return r.db.Table("product_doc_versions").Save(&v).Error
}
func (r *productDocsRepository) PutGroup(v *productdomain.DocGroup) error {
	return r.db.Table("product_doc_groups").Save(v).Error
}
func (r *productDocsRepository) PutDocument(v *productdomain.ProductDocument) error {
	return r.db.Table("product_documents").Save(v).Error
}
func (r *productDocsRepository) AddRevision(v *productdomain.DocumentRevision) error {
	return r.db.Table("product_document_revisions").Create(v).Error
}
func (r *productDocsRepository) DeleteVersion(id string) error {
	return r.db.Table("product_doc_versions").Where("id = ?", id).Delete(&productdomain.DocVersion{}).Error
}
func (r *productDocsRepository) DeleteGroup(id uint64) error {
	return r.db.Table("product_doc_groups").Where("id = ?", id).Delete(&productdomain.DocGroup{}).Error
}
func (r *productDocsRepository) DeleteDocument(id uint64) error {
	return r.db.Table("product_documents").Where("id = ?", id).Delete(&productdomain.ProductDocument{}).Error
}
