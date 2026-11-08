package schema

import (
	settingsmodel "sitehub/common/settings/model"

	"gorm.io/gorm"
)

func Apply(db *gorm.DB) error { return db.AutoMigrate(&settingsmodel.SystemConfig{}) }
