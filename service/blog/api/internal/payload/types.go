package payload

import (
	"encoding/json"
	"strconv"
	"time"
)

type Uint64 uint64

func (value Uint64) MarshalJSON() ([]byte, error) {
	return json.Marshal(strconv.FormatUint(uint64(value), 10))
}

type Int64 int64

func (value Int64) MarshalJSON() ([]byte, error) {
	return json.Marshal(strconv.FormatInt(int64(value), 10))
}

type List[T any] []T

func (value List[T]) MarshalJSON() ([]byte, error) {
	if value == nil {
		return []byte("[]"), nil
	}
	return json.Marshal([]T(value))
}

type ArticleHistory struct {
	ID        Uint64     `json:"id"`
	Version   Uint64     `json:"version"`
	ActorID   Uint64     `json:"actor_id"`
	Action    string     `json:"action"`
	CreatedAt *time.Time `json:"created_at"`
}

type ArticleHistoryList struct {
	Items List[ArticleHistory] `json:"items"`
}

type ArticleLabel struct {
	ID   Uint64 `json:"id"`
	Name string `json:"name"`
}

type ArticleOptions struct {
	Categories List[ArticleLabel] `json:"categories"`
	Tags       List[ArticleLabel] `json:"tags"`
}

type ArticlePage struct {
	Items    List[ArticleView] `json:"items"`
	Total    Int64             `json:"total"`
	Page     Int64             `json:"page"`
	PageSize Int64             `json:"pageSize"`
}

type ArticleStats struct {
	Total      Int64       `json:"total"`
	Published  Int64       `json:"published"`
	Draft      Int64       `json:"draft"`
	TotalViews Int64       `json:"totalViews"`
	Trend      List[Int64] `json:"trend"`
}

type ArticleTrendData struct {
	Dates  List[string] `json:"dates"`
	Counts List[Int64]  `json:"counts"`
}

type ArticleView struct {
	ID                    Uint64             `json:"id"`
	Version               Uint64             `json:"version"`
	Title                 string             `json:"title"`
	Slug                  string             `json:"slug"`
	Summary               string             `json:"summary"`
	Content               string             `json:"content"`
	ContentFormat         string             `json:"content_format"`
	Cover                 string             `json:"cover"`
	CategoryID            Uint64             `json:"category_id"`
	TagIDs                List[Uint64]       `json:"tag_ids"`
	Status                Uint64             `json:"status"`
	IsTop                 bool               `json:"is_top"`
	Author                *ArticleLabel      `json:"author"`
	Category              *ArticleLabel      `json:"category"`
	Tags                  List[ArticleLabel] `json:"tags"`
	CreatedBy             Uint64             `json:"created_by"`
	UpdatedBy             Uint64             `json:"updated_by"`
	CreatedAt             *time.Time         `json:"created_at"`
	UpdatedAt             *time.Time         `json:"updated_at"`
	PublishedAt           *time.Time         `json:"published_at"`
	ScheduledAt           *time.Time         `json:"scheduled_at"`
	DeletedAt             *time.Time         `json:"deleted_at"`
	Views                 Uint64             `json:"views"`
	HasUnpublishedChanges bool               `json:"has_unpublished_changes"`
}

type AssetInfo struct {
	ID   string `json:"id"`
	URL  string `json:"url"`
	Name string `json:"name"`
	MIME string `json:"mime"`
	Size Int64  `json:"size"`
}

type CategoryPage struct {
	Items List[Category] `json:"items"`
	Total Int64          `json:"total"`
}

type DashboardStats struct {
	Overview *DashboardOverview `json:"overview"`
	Trend    *DashboardTrend    `json:"trend"`
	Content  *DashboardContent  `json:"content"`
}

type ArticleRecord struct {
	Version          Uint64             `json:"version"`
	ContentFormat    string             `json:"content_format"`
	CreatedBy        Uint64             `json:"created_by"`
	UpdatedBy        Uint64             `json:"updated_by"`
	AuthorName       string             `json:"author_name"`
	PublishedVersion Uint64             `json:"published_version"`
	ScheduledVersion Uint64             `json:"scheduled_version"`
	DeletedAt        *time.Time         `json:"deleted_at"`
	ID               Uint64             `json:"id"`
	Title            string             `json:"title"`
	Slug             string             `json:"slug"`
	Summary          string             `json:"summary"`
	Content          string             `json:"content"`
	Cover            string             `json:"cover"`
	CategoryID       Uint64             `json:"category_id"`
	AuthorID         Uint64             `json:"author_id"`
	Views            Uint64             `json:"views"`
	Likes            Uint64             `json:"likes"`
	Favorites        Uint64             `json:"favorites"`
	Comments         Uint64             `json:"comments"`
	Status           Uint64             `json:"status"`
	IsTop            bool               `json:"is_top"`
	SortOrder        Int64              `json:"sort_order"`
	PublishedAt      *time.Time         `json:"published_at"`
	ScheduledAt      *time.Time         `json:"scheduled_at"`
	CreatedAt        *time.Time         `json:"created_at"`
	UpdatedAt        *time.Time         `json:"updated_at"`
	Category         *Category          `json:"category"`
	Tags             List[Tag]          `json:"tags"`
	ArticleViews     List[ArticleVisit] `json:"article_views"`
}

type ArticleVisit struct {
	ID        Uint64         `json:"id"`
	ArticleID Uint64         `json:"article_id"`
	IPAddress string         `json:"ip_address"`
	UserAgent string         `json:"user_agent"`
	UserID    *Uint64        `json:"user_id,omitempty"`
	CreatedAt *time.Time     `json:"created_at"`
	Article   *ArticleRecord `json:"article"`
	User      *PublicUser    `json:"user"`
}

type Category struct {
	ParentID     *Uint64             `json:"parent_id,omitempty"`
	ID           Uint64              `json:"id"`
	Name         string              `json:"name"`
	Slug         string              `json:"slug"`
	Description  string              `json:"description"`
	Icon         string              `json:"icon"`
	SortOrder    Int64               `json:"sort_order"`
	Status       Uint64              `json:"status"`
	ArticleCount Int64               `json:"article_count"`
	CreatedAt    *time.Time          `json:"created_at"`
	UpdatedAt    *time.Time          `json:"updated_at"`
	Articles     List[ArticleRecord] `json:"articles"`
}

type Tag struct {
	ID           Uint64              `json:"id"`
	Name         string              `json:"name"`
	Slug         string              `json:"slug"`
	Description  string              `json:"description"`
	ArticleCount Int64               `json:"article_count"`
	CreatedAt    *time.Time          `json:"created_at"`
	Articles     List[ArticleRecord] `json:"articles"`
}

type PublicUser struct {
	ID        Uint64     `json:"id"`
	Username  string     `json:"username"`
	Email     string     `json:"email"`
	Nickname  string     `json:"nickname"`
	Avatar    string     `json:"avatar"`
	Bio       string     `json:"bio"`
	Role      Uint64     `json:"role"`
	Status    Uint64     `json:"status"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

type FileInfo struct {
	URL        string `json:"url"`
	Name       string `json:"name"`
	MIMEType   string `json:"mimeType"`
	Size       Int64  `json:"size"`
	ModifiedAt Int64  `json:"modifiedAt"`
}

type DashboardOverview struct {
	Articles Int64 `json:"articles"`
	Views    Int64 `json:"views"`
	Likes    Int64 `json:"likes"`
	Comments Int64 `json:"comments"`
}

type DashboardTrend struct {
	Dates    List[string] `json:"dates"`
	Views    List[Int64]  `json:"views"`
	Articles List[Int64]  `json:"articles"`
}

type DashboardContent struct {
	Categories Int64 `json:"categories"`
	Tags       Int64 `json:"tags"`
}

type StatisticsRows struct {
	Items List[map[string]any] `json:"items"`
}

type TagPage struct {
	Items List[Tag] `json:"items"`
	Total Int64     `json:"total"`
}

type ViewTrendData struct {
	Dates List[string] `json:"dates"`
	Views List[Int64]  `json:"views"`
}

type WebsiteStorage struct {
	Provider    string       `json:"provider"`
	LocalPath   string       `json:"localPath"`
	URLPrefix   string       `json:"urlPrefix"`
	CDN         string       `json:"cdn"`
	Endpoint    string       `json:"endpoint"`
	AccessKey   string       `json:"accessKey"`
	SecretKey   string       `json:"secretKey"`
	Bucket      string       `json:"bucket"`
	Region      string       `json:"region"`
	PathStyle   bool         `json:"pathStyle"`
	AllowedExts List[string] `json:"allowedExts"`
	MaxSize     Int64        `json:"maxSize"`
}
