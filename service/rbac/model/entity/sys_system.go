package entity

// SysSystem identifies an independently routed frontend and its menu boundary.
type SysSystem struct {
	BaseModel
	Code        string `gorm:"size:50;not null;uniqueIndex" json:"code"`
	Name        string `gorm:"size:100;not null" json:"name"`
	Description string `gorm:"size:500" json:"description"`
	Origin      string `gorm:"size:255" json:"origin"`
	Enabled     bool   `gorm:"not null" json:"enabled"`
	Public      bool   `gorm:"not null" json:"public"`
}

func (SysSystem) TableName() string { return "sys_system" }
