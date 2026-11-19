package business

import (
	"errors"
	"testing"

	blogrepository "sitehub/service/blog/model"
)

type failingStatisticsRepository struct {
	blogrepository.
		StatisticsRepository
	err error
}

func (r *failingStatisticsRepository) GetArticleCount(*int) (int64, error)   { return 0, r.err }
func (r *failingStatisticsRepository) GetArticleTrend(int) ([]int, error)    { return nil, r.err }
func (r *failingStatisticsRepository) GetViewsTrendByDay(int) ([]int, error) { return nil, r.err }

func TestStatisticsFailureIsNotReportedAsZeroActivity(t *testing.T) {
	failure := errors.New("database unavailable")
	service := NewStatisticsService(&failingStatisticsRepository{err: failure})
	if _, err := service.GetDashboardStats(); !errors.Is(err, failure) {
		t.Fatalf("dashboard: %v", err)
	}
	if _, err := service.GetArticleStats(); !errors.Is(err, failure) {
		t.Fatalf("articles: %v", err)
	}
	if _, err := service.GetArticleTrend(7); !errors.Is(err, failure) {
		t.Fatalf("article trend: %v", err)
	}
	if _, err := service.GetViewTrend(7); !errors.Is(err, failure) {
		t.Fatalf("view trend: %v", err)
	}
}
