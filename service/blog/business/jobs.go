package services

import (
	"context"
	"time"
)

// PublishScheduledArticles keeps publication locks, snapshots and audit writes
// inside blog. Scheduler only triggers the operation through this capability.
func (s *articleService) PublishScheduledArticles(ctx context.Context) (int64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	worker := &articleService{articleRepo: s.articleRepo.WithContext(ctx)}
	return worker.publishDue(time.Now())
}
