package domain

import (
	"errors"
	"time"
)

var ErrDocumentNotFound = errors.New("文档或版本不存在")
var ErrDocumentConflict = errors.New("文档数据已变化，请重新加载后再操作；当前编辑内容已保留")
var ErrInvalidDocument = errors.New("文档参数无效")

// 产品版本与文档修订是两种不同的版本；正文始终为 Markdown 源码。
type DocSite struct {
	ID                                 uint64
	Title, Description, DefaultVersion string
	Revision                           uint64
}
type DocVersion struct {
	ID                   string `gorm:"primaryKey;size:80"`
	Name                 string `gorm:"size:60;not null"`
	Status               string `gorm:"size:20;not null"`
	ModeID               string `gorm:"size:80;not null;index:idx_doc_version_mode"`
	Revision             uint64
	CreatedAt, UpdatedAt time.Time
}
type DocMode struct {
	ID             string `gorm:"primaryKey;size:80"`
	Label          string `gorm:"size:60;not null"`
	Description    string `gorm:"size:500"`
	Icon           string `gorm:"size:100"`
	Status         string `gorm:"size:20;not null"`
	DefaultVersion string `gorm:"size:80"`
	SortOrder      int
	Revision       uint64
}
type DocModeInput struct {
	Action, ID, Label, Description, Icon, Status, DefaultVersion string
	SortOrder                                                    int
	Revision                                                     uint64
}
type DocGroup struct {
	ID        uint64
	VersionID string `gorm:"size:80;not null;index:idx_doc_group_version"`
	Title     string `gorm:"size:60;not null"`
	SortOrder int
	Revision  uint64
}
type ProductDocument struct {
	ID                   uint64
	VersionID            string `gorm:"size:80;not null;uniqueIndex:uk_document_version_key;uniqueIndex:uk_document_version_slug"`
	GroupID              uint64 `gorm:"not null;index:idx_document_group"`
	DocKey               string `gorm:"size:200;not null;uniqueIndex:uk_document_version_key"`
	Slug                 string `gorm:"size:200;not null;uniqueIndex:uk_document_version_slug"`
	Title                string `gorm:"size:200;not null"`
	Summary              string `gorm:"type:text"`
	ContentMarkdown      string `gorm:"type:longtext;not null"`
	SortOrder            int
	Revision, UpdatedBy  uint64
	CreatedAt, UpdatedAt time.Time
}
type DocumentRevision struct {
	ID              uint64    `json:"id"`
	DocumentID      uint64    `json:"documentId" gorm:"not null;uniqueIndex:uk_document_revision"`
	Revision        uint64    `json:"revision" gorm:"not null;uniqueIndex:uk_document_revision"`
	ActorID         uint64    `json:"actorId"`
	VersionID       string    `json:"versionId"`
	DocKey          string    `json:"docKey"`
	Slug            string    `json:"slug"`
	Title           string    `json:"title"`
	Summary         string    `json:"summary"`
	ContentMarkdown string    `json:"contentMarkdown,omitempty" gorm:"type:longtext"`
	GroupTitle      string    `json:"groupTitle"`
	Operation       string    `json:"operation"`
	CreatedAt       time.Time `json:"createdAt"`
}

type DocSettingsInput struct {
	Title, Description string
	Revision           uint64
}
type DocVersionInput struct {
	Action, ID, Name, Status, Source string
	ModeID                           string
	Revision                         uint64
}
type DocGroupInput struct {
	Action, Version, Title string
	ID, Revision           uint64
	Direction              int
}
type DocumentInput struct {
	Version, ID, Slug, Title, Summary, Content string
	GroupID, Revision                          uint64
}
type DocumentDeleteInput struct {
	Version, ID string
	Revision    uint64
}
