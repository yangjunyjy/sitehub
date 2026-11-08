package response

// DashboardStats 仪表盘统计数据
type DashboardStats struct {
	Overview struct {
		Articles int `json:"articles"`
		Views    int `json:"views"`
		Likes    int `json:"likes"`
		Comments int `json:"comments"`
	} `json:"overview"`
	Trend struct {
		Dates    []string `json:"dates"`
		Views    []int    `json:"views"`
		Articles []int    `json:"articles"`
	} `json:"trend"`
	Content struct {
		Categories int `json:"categories"`
		Tags       int `json:"tags"`
	} `json:"content"`
}

// ArticleStats 文章统计
type ArticleStats struct {
	Total      int   `json:"total"`
	Published  int   `json:"published"`
	Draft      int   `json:"draft"`
	TotalViews int   `json:"totalViews"`
	Trend      []int `json:"trend"`
}

// ViewTrendData 浏览量趋势数据（带日期）
type ViewTrendData struct {
	Dates  []string `json:"dates"`
	Values []int    `json:"views"`
}

// ArticleTrendData 文章发布趋势数据（带日期）
type ArticleTrendData struct {
	Dates  []string `json:"dates"`
	Values []int    `json:"counts"`
}
