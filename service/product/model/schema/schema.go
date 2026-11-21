package schema

import (
	settingsmodel "sitehub/common/settings/model"
	domain "sitehub/service/product/domain"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func Apply(db *gorm.DB) error {
	if err := ApplyReleases(db); err != nil {
		return err
	}
	if err := db.AutoMigrate(&settingsmodel.SystemConfig{}); err != nil {
		return err
	}
	for _, item := range []struct {
		table string
		model any
	}{{"product_doc_sites", &domain.DocSite{}}, {"product_doc_modes", &domain.DocMode{}}, {"product_doc_versions", &domain.DocVersion{}}, {"product_doc_groups", &domain.DocGroup{}}, {"product_documents", &domain.ProductDocument{}}, {"product_document_revisions", &domain.DocumentRevision{}}} {
		if err := db.Table(item.table).AutoMigrate(item.model); err != nil {
			return err
		}
	}
	return db.Transaction(func(tx *gorm.DB) error {
		for _, item := range []struct {
			table string
			model any
		}{{"product_doc_sites", &domain.DocSite{ID: 1, Title: "产品文档", DefaultVersion: "v1", Revision: 1}}, {"product_doc_modes", &domain.DocMode{ID: "guide", Label: "使用指南", Status: "published", DefaultVersion: "v1", Revision: 1}}, {"product_doc_versions", &domain.DocVersion{ID: "v1", Name: "v1", Status: "published", ModeID: "guide", Revision: 1}}} {
			if err := tx.Table(item.table).Clauses(clause.OnConflict{DoNothing: true}).Create(item.model).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
