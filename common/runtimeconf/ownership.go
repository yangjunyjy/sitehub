package runtimeconf

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

func SchemaVersion(service string) int {
	if service == "product" {
		return 2
	}
	if service == "blog" {
		return 6
	}
	if service == "rbac" {
		return 4
	}
	return 1
}

// RequireOwner prevents a service from starting against another domain's database.
func RequireOwner(ctx context.Context, db *gorm.DB, service string) error {
	var row struct {
		Service string
		Version int
	}
	if err := db.WithContext(ctx).Table("microservice_schema").Where("id = 1").Take(&row).Error; err != nil {
		return fmt.Errorf("initialize the %s database with tools/bootstrap: %w", service, err)
	}
	version := SchemaVersion(service)
	if row.Service != service || row.Version != version {
		return fmt.Errorf("database schema owner/version mismatch: expected %s version %d", service, version)
	}
	return nil
}
