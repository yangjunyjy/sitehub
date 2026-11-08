package services

import (
	"context"
	"fmt"
	"sync"

	"github.com/robfig/cron/v3"
	"github.com/zeromicro/go-zero/core/logx"

	"sitehub/service/scheduler/cronjob/internal/config"
)

// These scheduler-owned ports can become RPC clients later. Business
// repositories and transactions always remain inside their owning module.
type Publisher interface {
	PublishScheduledArticles(context.Context) (int64, error)
}
type Cleaner interface{ CleanupArticleAssets(context.Context) error }
type SchedulerService interface {
	Start()
	Stop()
	PublishScheduledArticles() (int64, error)
}
type schedulerService struct {
	publisher Publisher
	cleaner   Cleaner
	cron      *cron.Cron
	ctx       context.Context
	cancel    context.CancelFunc
	start     sync.Once
	stop      sync.Once
	lifecycle sync.Mutex
	config    config.Config
}

func NewSchedulerService(publisher Publisher, cleaner Cleaner) SchedulerService {
	s, err := NewWithConfig(publisher, cleaner, config.Config{})
	if err != nil {
		panic(err)
	}
	return s
}

func NewWithConfig(publisher Publisher, cleaner Cleaner, c config.Config) (SchedulerService, error) {
	c, err := c.Normalize()
	if err != nil {
		return nil, err
	}
	if publisher == nil {
		return nil, fmt.Errorf("scheduler publisher is required")
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &schedulerService{publisher: publisher, cleaner: cleaner, ctx: ctx, cancel: cancel, config: c,
		cron: cron.New(cron.WithSeconds(), cron.WithChain(cron.Recover(cron.DefaultLogger), cron.SkipIfStillRunning(cron.DefaultLogger)))}, nil
}
func (s *schedulerService) Start() {
	s.lifecycle.Lock()
	defer s.lifecycle.Unlock()
	s.start.Do(func() {
		if s.ctx.Err() != nil {
			return
		}
		s.add(s.config.PublishSchedule, "blog.publish", func(ctx context.Context) error { _, err := s.publisher.PublishScheduledArticles(ctx); return err })
		if s.cleaner != nil {
			s.add(s.config.CleanupSchedule, "blog.assets.cleanup", s.cleaner.CleanupArticleAssets)
		}
		s.cron.Start()
	})
}
func (s *schedulerService) add(schedule, name string, job func(context.Context) error) {
	_, err := s.cron.AddFunc(schedule, func() {
		ctx, cancel := context.WithTimeout(s.ctx, s.config.JobTimeout)
		defer cancel()
		if err := job(ctx); err != nil {
			logx.WithContext(ctx).Errorf("scheduler job %s failed: %v", name, err)
		}
	})
	if err != nil {
		panic(fmt.Errorf("register scheduler job %s: %w", name, err))
	}
}
func (s *schedulerService) Stop() {
	s.stop.Do(func() {
		s.lifecycle.Lock()
		s.cancel()
		drained := s.cron.Stop()
		s.lifecycle.Unlock()
		<-drained.Done()
	})
}
func (s *schedulerService) PublishScheduledArticles() (int64, error) {
	ctx, cancel := context.WithTimeout(s.ctx, s.config.JobTimeout)
	defer cancel()
	return s.publisher.PublishScheduledArticles(ctx)
}
