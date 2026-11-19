package business

import (
	"fmt"
	"time"

	blogresponse "sitehub/service/blog/dto/response"
	blogrepository "sitehub/service/blog/model"
)

// StatisticsService 统计服务接口
type StatisticsService interface {
	GetDashboardStats() (*blogresponse.DashboardStats, error)
	GetArticleStats() (*blogresponse.ArticleStats, error)
	GetArticleTrend(days int) (*blogresponse.ArticleTrendData, error)
	GetViewTrend(days int) (*blogresponse.ViewTrendData, error)
	GetHotArticles(limit int) ([]interface{}, error)
	GetCategoryStats() ([]interface{}, error)
	GetTagStats() ([]interface{}, error)
	GetHotTags(limit int) ([]interface{}, error)
}

// statisticsService 统计服务实现（私有结构体）
type statisticsService struct {
	statsRepo blogrepository.StatisticsRepository
}

// NewStatisticsService 创建统计服务（构造函数返回接口类型）
func NewStatisticsService(statsRepo blogrepository.StatisticsRepository) StatisticsService {
	return &statisticsService{
		statsRepo: statsRepo,
	}
}

// GetDashboardStats 获取仪表盘统计数据
func (s *statisticsService) GetDashboardStats() (*blogresponse.DashboardStats, error) {
	stats := &blogresponse.DashboardStats{}

	// 获取文章总数
	articleCount, err := s.statsRepo.GetArticleCount(nil)
	if err != nil {
		return nil, fmt.Errorf("读取文章数量: %w", err)
	}
	stats.Overview.Articles = int(articleCount)

	// 获取总浏览量
	totalViews, err := s.statsRepo.GetTotalViews()
	if err != nil {
		return nil, fmt.Errorf("读取访问数量: %w", err)
	}
	stats.Overview.Views = int(totalViews)

	// 获取总点赞数
	totalLikes, err := s.statsRepo.GetTotalLikes()
	if err != nil {
		return nil, fmt.Errorf("读取点赞数量: %w", err)
	}
	stats.Overview.Likes = int(totalLikes)

	// 获取总评论数
	commentCount, err := s.statsRepo.GetCommentCount(nil)
	if err != nil {
		return nil, fmt.Errorf("读取历史评论数量: %w", err)
	}
	stats.Overview.Comments = int(commentCount)

	// 获取当前内容模块的分类和标签数量
	categoryCount, err := s.statsRepo.GetCategoryCount()
	if err != nil {
		return nil, fmt.Errorf("读取分类数量: %w", err)
	}
	tagCount, err := s.statsRepo.GetTagCount()
	if err != nil {
		return nil, fmt.Errorf("读取标签数量: %w", err)
	}
	stats.Content.Categories = int(categoryCount)
	stats.Content.Tags = int(tagCount)

	// 获取趋势数据（包含日期）
	viewsTrend, err := s.GetViewTrend(7)
	if err != nil {
		return nil, err
	}
	articlesTrend, err := s.GetArticleTrend(7)
	if err != nil {
		return nil, err
	}
	stats.Trend.Views = viewsTrend.Values
	stats.Trend.Articles = articlesTrend.Values
	stats.Trend.Dates = viewsTrend.Dates

	return stats, nil
}

// GetArticleStats 获取文章统计数据
func (s *statisticsService) GetArticleStats() (*blogresponse.ArticleStats, error) {
	status1 := 1
	status0 := 0

	total, err := s.statsRepo.GetArticleCount(nil)
	if err != nil {
		return nil, fmt.Errorf("读取文章数量: %w", err)
	}
	published, err := s.statsRepo.GetArticleCount(&status1)
	if err != nil {
		return nil, fmt.Errorf("读取已发布数量: %w", err)
	}
	draft, err := s.statsRepo.GetArticleCount(&status0)
	if err != nil {
		return nil, fmt.Errorf("读取草稿数量: %w", err)
	}
	totalViews, err := s.statsRepo.GetTotalViews()
	if err != nil {
		return nil, fmt.Errorf("读取访问数量: %w", err)
	}

	trendData, err := s.GetArticleTrend(7)
	if err != nil {
		return nil, err
	}

	stats := &blogresponse.ArticleStats{
		Total:      int(total),
		Published:  int(published),
		Draft:      int(draft),
		TotalViews: int(totalViews),
		Trend:      trendData.Values,
	}
	return stats, nil
}

// GetArticleTrend 获取文章发布趋势（返回日期+数据）
func (s *statisticsService) GetArticleTrend(days int) (*blogresponse.ArticleTrendData, error) {
	trend, err := s.statsRepo.GetArticleTrend(days)
	if err != nil {
		return nil, fmt.Errorf("读取文章趋势: %w", err)
	}
	return &blogresponse.ArticleTrendData{Dates: generateDateLabels(days), Values: trend}, nil
}

// GetViewTrend 获取浏览量趋势（返回日期+数据）
func (s *statisticsService) GetViewTrend(days int) (*blogresponse.ViewTrendData, error) {
	trend, err := s.statsRepo.GetViewsTrendByDay(days)
	if err != nil {
		return nil, fmt.Errorf("读取访问趋势: %w", err)
	}
	return &blogresponse.ViewTrendData{Dates: generateDateLabels(days), Values: trend}, nil
}

// GetHotArticles 获取热门文章
func (s *statisticsService) GetHotArticles(limit int) ([]interface{}, error) {
	return s.statsRepo.GetHotArticles(limit)
}

// GetCategoryStats 获取分类统计
func (s *statisticsService) GetCategoryStats() ([]interface{}, error) {
	return s.statsRepo.GetCategoryStats()
}

// GetTagStats 获取标签统计
func (s *statisticsService) GetTagStats() ([]interface{}, error) {
	return s.statsRepo.GetTagStats()
}

// GetHotTags 获取热门标签
func (s *statisticsService) GetHotTags(limit int) ([]interface{}, error) {
	return s.statsRepo.GetHotTags(limit)
}

// generateDateLabels 生成最近 N 天的日期标签 (MM-DD 格式)
func generateDateLabels(days int) []string {
	labels := make([]string, days)
	for i := days - 1; i >= 0; i-- {
		labels[days-1-i] = time.Now().AddDate(0, 0, -i).Format("01-02")
	}
	return labels
}
