package response

import "time"

// CategoryItem 分类项
type CategoryItem struct {
	ID           uint      `json:"id"`
	Name         string    `json:"name"`
	Slug         string    `json:"slug"`
	Icon         string    `json:"icon"`
	SortOrder    int       `json:"sort_order"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	ArticleCount int64     `json:"articleCount"`
}

// CategoryDetailResponse 分类详情响应
type CategoryDetailResponse struct {
	Category CategoryInfo         `json:"category"`
	Articles *ArticleListResponse `json:"articles"`
}
