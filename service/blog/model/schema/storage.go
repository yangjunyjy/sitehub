package schema

import (
	settingsmodel "sitehub/common/settings/model"

	"gorm.io/gorm"
)

// ApplyStorage initializes the storage settings schema in the existing file-owned
// database. Blog owns this capability without implicitly migrating stored data.
func ApplyStorage(db *gorm.DB) error { return db.AutoMigrate(&settingsmodel.SystemConfig{}) }
