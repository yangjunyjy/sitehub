package services

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"sitehub/service/scheduler/cronjob/internal/config"
)

func TestConfiguredSchedulerRejectsInvalidScheduleAndHonorsDeadline(t *testing.T) {
	publisher := publisherFunc(func(ctx context.Context) (int64, error) {
		<-ctx.Done()
		return 0, ctx.Err()
	})
	if _, err := NewWithConfig(publisher, nil, config.Config{PublishSchedule: "invalid cron"}); err == nil {
		t.Fatal("invalid schedule accepted")
	}
	if _, err := NewWithConfig(publisher, nil, config.Config{JobTimeout: -time.Second}); err == nil {
		t.Fatal("negative timeout accepted")
	}
	s, err := NewWithConfig(publisher, nil, config.Config{JobTimeout: 10 * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Stop()
	if _, err := s.PublishScheduledArticles(); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("configured job budget ignored: %v", err)
	}
}

type publisherFunc func(context.Context) (int64, error)

func (f publisherFunc) PublishScheduledArticles(ctx context.Context) (int64, error) { return f(ctx) }

type cleanerFunc func(context.Context) error

func (f cleanerFunc) CleanupArticleAssets(ctx context.Context) error { return f(ctx) }

func TestSchedulerPropagatesBusinessFailureAndShutdown(t *testing.T) {
	cause := errors.New("publication unavailable")
	s := NewSchedulerService(publisherFunc(func(ctx context.Context) (int64, error) {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
		if _, ok := ctx.Deadline(); !ok {
			t.Error("job has no deadline")
		}
		return 3, cause
	}), nil)
	n, err := s.PublishScheduledArticles()
	if n != 3 || !errors.Is(err, cause) {
		t.Fatalf("result=%d err=%v", n, err)
	}
	s.Stop()
	if _, err = s.PublishScheduledArticles(); !errors.Is(err, context.Canceled) {
		t.Fatalf("job survived shutdown: %v", err)
	}
	s.Stop()
}

func TestSchedulerRegistersOnceSkipsOverlapAndCancelsRunningJob(t *testing.T) {
	entered := make(chan struct{})
	finished := make(chan struct{})
	var calls atomic.Int32
	s := NewSchedulerService(publisherFunc(func(ctx context.Context) (int64, error) {
		if calls.Add(1) == 1 {
			close(entered)
		}
		<-ctx.Done()
		return 0, ctx.Err()
	}), cleanerFunc(func(context.Context) error { return nil })).(*schedulerService)
	s.Start()
	s.Start()
	defer s.Stop()
	entries := s.cron.Entries()
	if len(entries) != 2 {
		t.Fatalf("duplicate job registration: %d", len(entries))
	}
	// Cron sorts Entries by the next execution time; the publication job was
	// registered first and must not be confused with cleanup near the hour.
	publication := s.cron.Entry(1)
	go func() { publication.WrappedJob.Run(); close(finished) }()
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("job did not start")
	}
	publication.WrappedJob.Run()
	if calls.Load() != 1 {
		t.Fatal("overlapping job ran twice")
	}
	s.Stop()
	select {
	case <-finished:
	case <-time.After(time.Second):
		t.Fatal("shutdown did not cancel job context")
	}
}
