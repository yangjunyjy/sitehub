package entity

import "gorm.io/gorm"

// Archived icons remain addressable for published references in other domains.
type SystemIcon struct {
	ID        uint64         `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"size:70;uniqueIndex;not null" json:"name"`
	Label     string         `gorm:"size:80;not null" json:"label"`
	Category  string         `gorm:"size:40;not null" json:"category"`
	SVG       string         `gorm:"type:mediumtext;not null" json:"svg,omitempty"`
	DeletedAt gorm.DeletedAt `json:"-"`
}
