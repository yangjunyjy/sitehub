package domain

import "sitehub/service/product/contract"

type ReleaseGroup struct {
	ID          string `json:"id" gorm:"primaryKey;size:120"`
	Name        string `json:"name" gorm:"size:80"`
	Description string `json:"description" gorm:"type:text"`
	SortOrder   int    `json:"sortOrder" gorm:"index"`
	Revision    uint64 `json:"revision,string" gorm:"not null"`
}

type ReleaseFile struct {
	ID          string `json:"id" gorm:"primaryKey;size:120"`
	GroupID     string `json:"groupId" gorm:"size:120;index"`
	Product     string `json:"product" gorm:"size:120"`
	Edition     string `json:"edition" gorm:"size:80"`
	Version     string `json:"version" gorm:"size:120"`
	Platform    string `json:"platform" gorm:"size:80"`
	Arch        string `json:"arch" gorm:"size:120"`
	Format      string `json:"format"`
	URL         string `json:"url" gorm:"type:text"`
	FileName    string `json:"fileName" gorm:"size:255"`
	Description string `json:"description" gorm:"type:text"`
	Icon        string `json:"icon" gorm:"size:255"`
	Checksum    string `json:"checksum" gorm:"size:64"`
	ReleaseDate string `json:"releaseDate"`
	Size        int64  `json:"size"`
	SortOrder   int    `json:"sortOrder" gorm:"index"`
	Revision    uint64 `json:"revision,string" gorm:"not null"`
	Status      string `json:"status" gorm:"size:16;index"`
}

type ReleaseSettings struct {
	ID                   uint                            `gorm:"primaryKey;autoIncrement:false" json:"-"`
	Title                string                          `json:"title"`
	Description          string                          `json:"description" gorm:"type:text"`
	Version              string                          `json:"version" gorm:"size:120"`
	ReleaseDate          string                          `json:"releaseDate"`
	SupportArchitectures string                          `json:"supportArchitectures"`
	SupportTitle         string                          `json:"supportTitle"`
	SupportDescription   string                          `json:"supportDescription"`
	InstallOptions       []contract.WebsiteInstallOption `json:"installOptions" gorm:"-"`
	Requirements         []contract.WebsiteRequirement   `json:"requirements" gorm:"-"`
	Revision             uint64                          `json:"revision,string" gorm:"not null"`
}

func (ReleaseGroup) TableName() string         { return "product_release_groups" }
func (ReleaseFile) TableName() string          { return "product_release_files" }
func (ReleaseSettings) TableName() string      { return "product_release_settings" }
func (ReleaseInstallOption) TableName() string { return "product_release_install_options" }
func (ReleaseRequirement) TableName() string   { return "product_release_requirements" }

type ReleaseInstallOption struct {
	Position                      int `gorm:"primaryKey;autoIncrement:false"`
	contract.WebsiteInstallOption `gorm:"embedded"`
}
type ReleaseRequirement struct {
	Position                    int `gorm:"primaryKey;autoIncrement:false"`
	contract.WebsiteRequirement `gorm:"embedded"`
}
type ReleaseCatalog struct {
	ReleaseSettings
	Groups   []ReleaseGroup `json:"groups"`
	Packages []ReleaseFile  `json:"packages"`
}
