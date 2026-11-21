package main

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"gorm.io/gorm"
	"sitehub/common/runtimeconf"
	services "sitehub/service/rbac/business"
	"sitehub/service/rbac/model/entity"
)

func TestIconCatalogUpgradeAndPublishedReferences(t *testing.T) {
	path := filepath.Join(t.TempDir(), "rbac.db")
	t.Setenv("RBAC_DB_DSN", path)
	if err := run("rbac", "sqlite", false); err != nil {
		t.Fatal(err)
	}
	db, err := (runtimeconf.Database{Driver: "sqlite", DSN: path, MaxOpen: 1}).Open()
	if err != nil {
		t.Fatal(err)
	}
	pool, _ := db.DB()
	defer pool.Close()
	// Reproduce the previous version's friend menu and a non-superadmin role grant.
	for _, stmt := range []string{
		"DROP TABLE system_icons",
		"UPDATE microservice_schema SET version=3",
		"DELETE FROM sys_menu WHERE component IN ('blog/comments/index','blog/messages/index')",
		"INSERT INTO sys_role_menu(role_id,menu_id,created_at) SELECT 92,id,CURRENT_TIMESTAMP FROM sys_menu WHERE component='blog/friends/index'",
	} {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 2; i++ {
		if err := run("rbac", "sqlite", false); err != nil {
			t.Fatal(err)
		}
	}
	var grants int64
	if err = db.Model(&entity.SysRoleMenu{}).Where("role_id = ?", 92).Count(&grants).Error; err != nil || grants != 3 {
		t.Fatalf("review grants=%d err=%v", grants, err)
	}
	ctx := context.Background()
	icons := services.NewSystemIcons(db)
	input := services.SystemIconInput{Name: "brand", SVG: `<svg viewBox="0 0 24 24"><circle r="8"/></svg>`}
	icon, err := icons.Create(ctx, input)
	if err != nil || icon.Name != "custom:brand" {
		t.Fatalf("create: %v, %v", icon, err)
	}
	if _, err := icons.Create(ctx, input); err == nil {
		t.Fatal("duplicate name accepted")
	}
	input.Label = "Brand"
	if _, err := icons.Update(ctx, icon.ID, input); err != nil {
		t.Fatal(err)
	}
	input.Name = "different"
	if _, err := icons.Update(ctx, icon.ID, input); err == nil {
		t.Fatal("changed a published identifier")
	}
	input.SVG = "<svg><image/></svg>"
	if _, err := icons.Validate(input); err == nil {
		t.Fatal("unsupported SVG accepted")
	}
	if err := icons.Delete(ctx, icon.ID); err != nil {
		t.Fatal(err)
	}
	list, err := icons.List(ctx)
	if err != nil || len(list) != 0 {
		t.Fatalf("picker after archive: %v %v", list, err)
	}
	content, err := icons.Content(ctx, icon.Name)
	if err != nil || content != icon.SVG {
		t.Fatalf("published SVG changed after archive: %v", err)
	}
	if err := icons.Delete(ctx, icon.ID); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("repeat archive: %v", err)
	}
}
