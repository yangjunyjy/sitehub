package logic

import "sitehub/service/blog/domain"

type articleHTTPView struct {
	domain.ArticleView
	ID         uint64 `json:"id,string"`
	Version    uint64 `json:"version,string"`
	CategoryID uint64 `json:"category_id,string"`
	CreatedBy  uint64 `json:"created_by,string,omitempty"`
	UpdatedBy  uint64 `json:"updated_by,string,omitempty"`
}

func articleHTTP(view domain.ArticleView) articleHTTPView {
	return articleHTTPView{ArticleView: view, ID: view.ID, Version: view.Version, CategoryID: view.CategoryID, CreatedBy: view.CreatedBy, UpdatedBy: view.UpdatedBy}
}
func articleResponse(value any) any {
	switch article := value.(type) {
	case domain.ArticleView:
		return articleHTTP(article)
	case domain.ArticlePage:
		items := make([]articleHTTPView, 0, len(article.Items))
		for _, entry := range article.Items {
			items = append(items, articleHTTP(entry))
		}
		return struct {
			domain.ArticlePage
			Items []articleHTTPView `json:"items"`
		}{article, items}
	default:
		return value
	}
}
