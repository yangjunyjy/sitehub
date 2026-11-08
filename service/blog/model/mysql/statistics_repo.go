package mysql

import (
	"time"

	"gorm.io/gorm"

	blogrepository "sitehub/service/blog/model"
	blogentity "sitehub/service/blog/model/entity"
)

type StatisticsRepositoryImpl struct {
	db *gorm.DB
}

func NewStatisticsRepository(db *gorm.DB) blogrepository.StatisticsRepository {
	return &StatisticsRepositoryImpl{db: db}
}

// GetArticleCount 获取文章数量
func (r *StatisticsRepositoryImpl) GetArticleCount(status *int) (int64, error) {
	query := r.db.Model(&blogentity.Article{}).Where("deleted_at IS NULL")
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	var count int64
	err := query.Count(&count).Error
	return count, err
}

// GetTotalViews 获取总浏览量
func (r *StatisticsRepositoryImpl) GetTotalViews() (int64, error) {
	var count int64
	err := r.db.Model(&blogentity.ArticleView{}).Count(&count).Error
	return count, err
}

// GetTotalLikes 获取总点赞数
func (r *StatisticsRepositoryImpl) GetTotalLikes() (int64, error) {
	var count int64
	err := r.db.Model(&blogentity.Article{}).Where("deleted_at IS NULL").Select("COALESCE(SUM(likes), 0)").Scan(&count).Error
	return count, err
}

// GetCommentCount 获取评论数量
func (r *StatisticsRepositoryImpl) GetCommentCount(status *int) (int64, error) {
	var count int64
	err := r.db.Model(&blogentity.Article{}).Where("deleted_at IS NULL").Select("COALESCE(SUM(comment_count), 0)").Scan(&count).Error
	return count, err
}

// GetCategoryCount 获取分类数量
func (r *StatisticsRepositoryImpl) GetCategoryCount() (int64, error) {
	var count int64
	err := r.db.Model(&blogentity.Category{}).Count(&count).Error
	return count, err
}

// GetTagCount 获取标签数量
func (r *StatisticsRepositoryImpl) GetTagCount() (int64, error) {
	var count int64
	err := r.db.Model(&blogentity.Tag{}).Count(&count).Error
	return count, err
}

// GetArticleTrend 获取文章发布趋势
func (r *StatisticsRepositoryImpl) GetArticleTrend(days int) ([]int, error) {
	var trend []int
	for i := days - 1; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		var count int64
		err := r.db.Model(&blogentity.Article{}).Where("deleted_at IS NULL").Where("DATE(created_at) = ?", date).Count(&count).Error
		if err != nil {
			return nil, err
		}
		trend = append(trend, int(count))
	}
	return trend, nil
}

// GetViewsTrend 获取浏览量趋势（总数）
func (r *StatisticsRepositoryImpl) GetViewsTrend(days int) (int64, error) {
	startDate := time.Now().AddDate(0, 0, -days)
	var count int64
	err := r.db.Model(&blogentity.ArticleView{}).Where("created_at >= ?", startDate).Count(&count).Error
	return count, err
}

// GetViewsTrendByDay 获取按天统计的浏览量趋势
func (r *StatisticsRepositoryImpl) GetViewsTrendByDay(days int) ([]int, error) {
	var trend []int
	for i := days - 1; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		var count int64
		err := r.db.Model(&blogentity.ArticleView{}).Where("DATE(created_at) = ?", date).Count(&count).Error
		if err != nil {
			return nil, err
		}
		trend = append(trend, int(count))
	}
	return trend, nil
}

// GetHotArticles 获取热门文章
func (r *StatisticsRepositoryImpl) GetHotArticles(limit int) ([]interface{}, error) {
	type HotArticle struct {
		ID           uint   `json:"id"`
		Title        string `json:"title"`
		ViewCount    int    `json:"views"`
		LikeCount    int    `json:"likes"`
		CommentCount int    `json:"commentCount"`
	}

	var articles []blogentity.Article
	err := r.db.Where("deleted_at IS NULL").Order("views DESC").Limit(limit).Find(&articles).Error
	if err != nil {
		return nil, err
	}

	result := make([]interface{}, len(articles))
	for i, article := range articles {
		result[i] = HotArticle{
			ID:           uint(article.ID),
			Title:        article.Title,
			ViewCount:    int(article.Views),
			LikeCount:    int(article.Likes),
			CommentCount: int(article.CommentCount),
		}
	}
	return result, nil
}

// GetHotTags 获取热门标签
func (r *StatisticsRepositoryImpl) GetHotTags(limit int) ([]interface{}, error) {
	type HotTag struct {
		ID           uint   `json:"id"`
		Name         string `json:"name"`
		ArticleCount int    `json:"articleCount"`
	}

	var tags []HotTag
	err := r.db.Model(&blogentity.Tag{}).
		Select("tag.id, tag.name, COUNT(article_tag.article_id) as article_count").
		Joins("LEFT JOIN article_tag ON tag.id = article_tag.tag_id").
		Group("tag.id, tag.name").
		Order("article_count DESC").
		Limit(limit).
		Find(&tags).Error
	if err != nil {
		return nil, err
	}

	result := make([]interface{}, len(tags))
	for i, tag := range tags {
		result[i] = tag
	}
	return result, nil
}

// GetCategoryStats 获取分类统计
func (r *StatisticsRepositoryImpl) GetCategoryStats() ([]interface{}, error) {
	type CategoryStat struct {
		ID           uint   `json:"id"`
		Name         string `json:"name"`
		ArticleCount int    `json:"articleCount"`
	}

	var stats []CategoryStat
	err := r.db.Model(&blogentity.Category{}).
		Select("id, name, (SELECT COUNT(*) FROM articles WHERE category_id = categorie.id AND status = 1 AND deleted_at IS NULL) as article_count").
		Find(&stats).Error
	if err != nil {
		return nil, err
	}

	result := make([]interface{}, len(stats))
	for i, stat := range stats {
		result[i] = stat
	}
	return result, nil
}

// GetTagStats 获取标签统计
func (r *StatisticsRepositoryImpl) GetTagStats() ([]interface{}, error) {
	type TagStat struct {
		ID           uint   `json:"id"`
		Name         string `json:"name"`
		ArticleCount int    `json:"articleCount"`
	}

	var stats []TagStat
	err := r.db.Model(&blogentity.Tag{}).
		Select("tag.id, tag.name, COUNT(article_tag.article_id) as article_count").
		Joins("LEFT JOIN article_tag ON tag.id = article_tag.tag_id").
		Group("tag.id, tag.name").
		Find(&stats).Error
	if err != nil {
		return nil, err
	}

	result := make([]interface{}, len(stats))
	for i, stat := range stats {
		result[i] = stat
	}
	return result, nil
}
