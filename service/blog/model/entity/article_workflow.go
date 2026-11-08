package entity

import "time"

// Revision snapshots are immutable; restoring creates another revision.
type ArticleRevision struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	ArticleID uint64    `gorm:"uniqueIndex:article_revision_version" json:"article_id"`
	Version   uint64    `gorm:"uniqueIndex:article_revision_version" json:"version"`
	Payload   string    `gorm:"type:longtext" json:"-"`
	ActorID   uint64    `json:"actor_id"`
	Action    string    `gorm:"size:40" json:"action"`
	CreatedAt time.Time `json:"created_at"`
}

func (ArticleRevision) TableName() string { return "article_revisions" }

type ArticleOperation struct {
	ID        uint64 `gorm:"primaryKey"`
	ActorID   uint64 `gorm:"uniqueIndex:article_operation_key"`
	RequestID string `gorm:"size:80;uniqueIndex:article_operation_key"`
	Digest    string `gorm:"size:64"`
	Result    string `gorm:"type:longtext"`
	CreatedAt time.Time
}

func (ArticleOperation) TableName() string { return "article_operations" }

// Published content is independent of the editable articles row.
type ArticlePublication struct {
	ArticleID   uint64    `gorm:"primaryKey"`
	Title       string    `gorm:"size:200"`
	Slug        string    `gorm:"size:200"`
	Content     string    `gorm:"type:longtext"`
	CategoryID  uint64    `gorm:"index:article_publication_category"`
	Payload     string    `gorm:"type:longtext"`
	PublishedAt time.Time `gorm:"index:article_publication_time"`
}

func (ArticlePublication) TableName() string { return "article_publications" }

type ArticlePublicationTag struct {
	ArticleID uint64 `gorm:"primaryKey"`
	TagID     uint64 `gorm:"primaryKey;index:article_publication_tag"`
}

func (ArticlePublicationTag) TableName() string { return "article_publication_tags" }

type ArticleAsset struct {
	ID        string     `gorm:"primaryKey;size:36" json:"id"`
	OwnerID   uint64     `gorm:"index:article_asset_owner" json:"owner_id"`
	Name      string     `gorm:"size:255" json:"name"`
	MIME      string     `gorm:"size:120" json:"mime"`
	Size      int64      `json:"size"`
	SHA256    string     `gorm:"size:64" json:"sha256"`
	Location  string     `gorm:"type:text" json:"-"`
	Storage   string     `gorm:"type:text" json:"-"`
	CreatedAt time.Time  `json:"created_at"`
	DeletedAt *time.Time `gorm:"index:article_asset_deleted" json:"deleted_at"`
}

func (ArticleAsset) TableName() string { return "article_assets" }

type ArticleAssetReference struct {
	ArticleID uint64 `gorm:"primaryKey"`
	AssetID   string `gorm:"primaryKey;size:36"`
	Version   uint64 `gorm:"primaryKey"`
}

func (ArticleAssetReference) TableName() string { return "article_asset_references" }
