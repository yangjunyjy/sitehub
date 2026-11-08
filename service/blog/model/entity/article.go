package entity

import (
	"time"
)

// Article 文章模型，存储博客文章的基本信息和统计数据
type Article struct {
	Version          uint64     `gorm:"not null;default:1" json:"version"`
	ContentFormat    string     `gorm:"size:20;default:markdown" json:"content_format"`
	CreatedBy        uint64     `json:"created_by"`
	UpdatedBy        uint64     `json:"updated_by"`
	AuthorName       string     `gorm:"size:100" json:"author_name"`
	PublishedVersion uint64     `json:"published_version"`
	ScheduledVersion uint64     `json:"scheduled_version"`
	DeletedAt        *time.Time `gorm:"index" json:"deleted_at"`

	ID           uint64     `gorm:"primaryKey;autoIncrement" json:"id"`                                 // 文章唯一标识ID
	Title        string     `gorm:"size:200;not null" json:"title" validate:"required"`                 // 文章标题，必填
	Slug         string     `gorm:"size:200;not null;uniqueIndex" json:"slug" validate:"required,slug"` // 文章URL别名，用于生成友好的URL，全局唯一
	Summary      string     `gorm:"type:text" json:"summary"`                                           // 文章摘要，用于列表页展示
	Content      string     `gorm:"type:longtext;not null" json:"content" validate:"required"`          // 文章正文内容，使用长文本类型存储
	Cover        string     `gorm:"size:500" json:"cover"`                                              // 文章封面图片URL
	CategoryID   uint64     `gorm:"index" json:"category_id"`                                           // 文章所属分类的ID
	AuthorID     uint64     `gorm:"not null;index" json:"author_id"`                                    // 文章作者的ID
	Views        uint64     `gorm:"not null;default:0" json:"views"`                                    // 文章浏览次数统计
	Likes        uint32     `gorm:"not null;default:0" json:"likes"`                                    // 文章获得的点赞数量
	Favorites    uint32     `gorm:"not null;default:0" json:"favorites"`                                // 文章被收藏的次数
	CommentCount uint32     `gorm:"not null;default:0;column:comment_count" json:"comments"`            // 文章评论数量统计
	Status       uint8      `gorm:"not null;default:0" json:"status"`                                   // 默认为草稿，避免 GORM 将显式 0 替换为已发布
	IsTop        bool       `gorm:"not null;default:false" json:"is_top"`                               // 是否置顶显示，true表示置顶
	SortOrder    int        `gorm:"not null;default:0" json:"sort_order"`                               // 文章排序字段，数值越小越靠前
	PublishedAt  *time.Time `json:"published_at"`                                                       // 文章发布时间，用于按发布时间排序
	ScheduledAt  *time.Time `gorm:"index" json:"scheduled_at"`                                          // 定时发布时间，到达后自动变为已发布
	CreatedAt    time.Time  `gorm:"autoCreateTime" json:"created_at"`                                   // 文章创建时间，自动记录
	UpdatedAt    time.Time  `gorm:"autoUpdateTime" json:"updated_at"`                                   // 文章最后更新时间，自动记录

	// 关联关系（不使用外键约束）
	Category     *Category     `gorm:"foreignKey:CategoryID;references:ID" json:"category,omitempty"` // 文章所属分类
	Author       *User         `gorm:"-" json:"-"`                                                    // 文章作者信息
	Tags         []Tag         `gorm:"many2many:article_tag;" json:"tags,omitempty"`                  // 文章关联的标签列表（保持多对多关系）
	ArticleViews []ArticleView `gorm:"-" json:"article_views,omitempty"`                              // 文章的浏览记录列表
}

func (Article) TableName() string {
	return "articles"
}
