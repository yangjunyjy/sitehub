package response

import "time"

// TagItem 标签项
type TagItem struct {
	ID           uint      `json:"id"`
	Name         string    `json:"name"`
	Slug         string    `json:"slug"`
	CreatedAt    time.Time `json:"createdAt"`
	ArticleCount int64     `json:"articleCount"`
	Description  string    `json:"description"`
}

// TagDetailResponse 标签详情响应
type TagDetailResponse struct {
	Tag      TagInfo              `json:"tag"`
	Articles *ArticleListResponse `json:"articles"`
}
