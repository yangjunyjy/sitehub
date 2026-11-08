package entity

import (
	"time"

	"gorm.io/gorm"
)

type IPBlacklist struct {
	ID        uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	IPAddress string         `gorm:"size:45;not null;uniqueIndex:idx_ip" json:"ip_address"`
	Reason    string         `gorm:"type:text;not null" json:"reason"`
	IsBlocked bool           `gorm:"not null;default:true" json:"is_blocked"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index:ip_blacklist_idx_deleted_at" json:"-"`
}

func (IPBlacklist) TableName() string {
	return "ip_blacklist"
}
