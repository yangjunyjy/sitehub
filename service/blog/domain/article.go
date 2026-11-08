package domain

import (
	"errors"
	"time"
)

var (
	ErrArticleReferenced = errors.New("分类或标签仍被文章、线上版本或子分类引用，请先调整引用后再删除")
	ErrInvalidArticle    = errors.New("文章参数无效")
	ErrArticleConflict   = errors.New("文章已被其他操作修改，请重新加载并比较本地内容")
	ErrArticleSlug       = errors.New("URL 别名已存在，请修改后重试")
	ErrArticleReference  = errors.New("分类、标签或资源不存在或不可用")
	ErrArticleForbidden  = errors.New("没有执行该文章操作的权限")
	ErrArticleNotFound   = errors.New("文章或版本不存在")
	ErrArticleState      = errors.New("当前文章状态不允许此操作")
)

type ArticleActor struct {
	ID                                                             uint64
	Name                                                           string
	CanCreate, CanRead, CanUpdate, CanPublish, CanAudit, CanDelete bool
}
type ArticleInput struct {
	Title         string `json:"title"`
	Slug          string `json:"slug"`
	Summary       string `json:"summary"`
	Content       string `json:"content"`
	ContentFormat string `json:"content_format"`
	Cover         string `json:"cover"`
	CategoryID    uint64 `json:"category_id"`
	TagIDs        []uint `json:"tag_ids"`
}
type SaveArticle struct {
	ID        uint64
	Version   uint64
	RequestID string
	Input     ArticleInput
}
type ArticleAction struct {
	ID          uint64
	Version     uint64
	RequestID   string
	Action      string
	ScheduledAt *time.Time
	RevisionID  uint64
	IsTop       *bool
}
type ArticleLabel struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}
type ArticleView struct {
	ID      uint64 `json:"id"`
	Version uint64 `json:"version"`
	ArticleInput
	Status                uint8          `json:"status"`
	IsTop                 bool           `json:"is_top"`
	Author                ArticleLabel   `json:"author"`
	Category              *ArticleLabel  `json:"category"`
	Tags                  []ArticleLabel `json:"tags"`
	CreatedBy             uint64         `json:"created_by,omitempty"`
	UpdatedBy             uint64         `json:"updated_by,omitempty"`
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`
	PublishedAt           *time.Time     `json:"published_at"`
	ScheduledAt           *time.Time     `json:"scheduled_at"`
	DeletedAt             *time.Time     `json:"deleted_at"`
	Views                 uint64         `json:"views"`
	HasUnpublishedChanges bool           `json:"has_unpublished_changes"`
}
type ArticleListQuery struct {
	Page, PageSize int
	Title, Sort    string
	Status         *int
	CategoryID     uint64
	TagIDs         []uint
	Trash          bool
}
type ArticlePage struct {
	Items    []ArticleView `json:"items"`
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"pageSize"`
}
type ArticleHistory struct {
	ID        uint64    `json:"id"`
	Version   uint64    `json:"version"`
	ActorID   uint64    `json:"actor_id"`
	Action    string    `json:"action"`
	CreatedAt time.Time `json:"created_at"`
}
type ArticleOptions struct {
	Categories []ArticleLabel `json:"categories"`
	Tags       []ArticleLabel `json:"tags"`
}
