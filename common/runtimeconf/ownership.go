package runtimeconf

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// RequireOwner prevents a service from starting against another domain's database.
func RequireOwner(ctx context.Context, db *gorm.DB, service string) error {
	var row struct {
		Service string
		Version int
	}
	if err := db.WithContext(ctx).Table("microservice_schema").Where("id = 1").Take(&row).Error; err != nil {
		return fmt.Errorf("initialize the %s database with tools/bootstrap: %w", service, err)
	}
	if row.Service != service || row.Version != 1 {
		return fmt.Errorf("database schema owner/version mismatch: expected %s version 1", service)
	}
	return nil
}
