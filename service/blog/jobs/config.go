// Package jobs runs publication and asset cleanup inside the blog service.
package jobs

import (
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
)

type Config struct {
	Enabled         bool          `json:",default=true"`
	PublishSchedule string        `json:",optional"`
	CleanupSchedule string        `json:",optional"`
	JobTimeout      time.Duration `json:",default=2m"`
}

func (c Config) Normalize() (Config, error) {
	if c.PublishSchedule == "" {
		c.PublishSchedule = "@every 1m"
	}
	if c.CleanupSchedule == "" {
		c.CleanupSchedule = "0 0 * * * *"
	}
	if c.JobTimeout == 0 {
		c.JobTimeout = 2 * time.Minute
	}
	if c.JobTimeout < 0 {
		return c, fmt.Errorf("scheduler job timeout must be positive")
	}
	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	for name, expression := range map[string]string{"publish": c.PublishSchedule, "cleanup": c.CleanupSchedule} {
		if _, err := parser.Parse(expression); err != nil {
			return c, fmt.Errorf("scheduler %s schedule: %w", name, err)
		}
	}
	return c, nil
}
