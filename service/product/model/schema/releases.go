package schema

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	domain "sitehub/service/product/domain"
)

func ApplyReleases(db *gorm.DB) error {
	if err := db.AutoMigrate(&domain.ReleaseSettings{}, &domain.ReleaseGroup{}, &domain.ReleaseFile{}, &domain.ReleaseInstallOption{}, &domain.ReleaseRequirement{}); err != nil {
		return err
	}
	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(&domain.ReleaseSettings{ID: 1, Title: "下载中心", Revision: 1}).Error
}
