package main

import (
	"context"
	"os"
	"path/filepath"
	"sitehub/common/runtimeconf"
	blogmysql "sitehub/service/blog/model/mysql"
	"strings"
	"testing"
)

func TestSeparateServiceSchemasAndOwnership(t *testing.T) {
	for _, domain := range []string{"rbac", "blog", "product", "file"} {
		t.Run(domain, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "service.db")
			t.Setenv(strings.ToUpper(domain)+"_DB_DSN", path)
			if err := run(domain, "sqlite", false); err != nil {
				t.Fatal(err)
			}
			if err := run(domain, "sqlite", false); err != nil {
				t.Fatal("bootstrap not repeatable", err)
			}
			db, err := (runtimeconf.Database{Driver: "sqlite", DSN: path, MaxOpen: 1}).Open()
			if err != nil {
				t.Fatal(err)
			}
			pool, _ := db.DB()
			defer pool.Close()
			tables, err := db.Migrator().GetTables()
			if err != nil {
				t.Fatal(err)
			}
			for _, table := range tables {
				if (domain != "rbac" && (strings.HasPrefix(table, "sys_") || strings.HasPrefix(table, "product_member"))) || (domain != "blog" && (table == "articles" || strings.HasPrefix(table, "article_"))) || (domain != "product" && strings.HasPrefix(table, "product_doc")) {
					t.Fatalf("%s contains another domain's table %s", domain, table)
				}
			}
			if domain == "file" {
				t.Setenv("BLOG_DB_DSN", path)
				if err := run("blog", "sqlite", false); err == nil {
					t.Fatal("database ownership was not enforced")
				}
			}
		})
	}
}
func TestBootstrapRejectsNonServiceDatabase(t *testing.T) {
	path := filepath.Join(t.TempDir(), "original.db")
	db, err := (runtimeconf.Database{Driver: "sqlite", DSN: path, MaxOpen: 1}).Open()
	if err != nil {
		t.Fatal(err)
	}
	if err = db.Exec("CREATE TABLE original_data(id INTEGER)").Error; err != nil {
		t.Fatal(err)
	}
	pool, _ := db.DB()
	pool.Close()
	t.Setenv("FILE_DB_DSN", path)
	if err = run("file", "sqlite", false); err == nil {
		t.Fatal("original database accepted")
	}
	if _, err = os.Stat(path); err != nil {
		t.Fatal(err)
	}
}

func TestExistingBlogAndMenuSettingsUpgrade(t *testing.T) {
	for _, domain := range []string{"blog", "rbac"} {
		t.Run(domain, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "v2.db")
			t.Setenv(strings.ToUpper(domain)+"_DB_DSN", path)
			if err := run(domain, "sqlite", false); err != nil {
				t.Fatal(err)
			}
			db, err := (runtimeconf.Database{Driver: "sqlite", DSN: path, MaxOpen: 1}).Open()
			if err != nil {
				t.Fatal(err)
			}
			pool, _ := db.DB()
			defer pool.Close()
			setup := []string{"UPDATE microservice_schema SET version = 2 WHERE id = 1"}
			if domain == "blog" {
				setup = append(setup, "ALTER TABLE author_profiles DROP COLUMN home_animation", "ALTER TABLE author_profiles DROP COLUMN home_verse", "INSERT INTO author_profiles (id, name) VALUES (1, 'Existing author')")
			} else {
				setup = append(setup, "ALTER TABLE sys_menu DROP COLUMN page_type", "ALTER TABLE sys_menu DROP COLUMN external_url", "INSERT INTO sys_menu (menu_name, menu_type, menu_code, path) VALUES ('Journal', 2, 'existing-page', '/journal')")
			}
			for _, query := range setup {
				if err = db.Exec(query).Error; err != nil {
					t.Fatal(err)
				}
			}
			if err = run(domain, "sqlite", false); err != nil {
				t.Fatal(err)
			}
			ctx := context.Background()
			if err = runtimeconf.RequireOwner(ctx, db, domain); err != nil {
				t.Fatal(err)
			}
			if domain == "blog" {
				repo := blogmysql.NewCommunityRepository(db)
				if err = repo.SaveHome(ctx, false, "山色入窗"); err != nil {
					t.Fatal(err)
				}
				profile, readErr := repo.Profile(ctx)
				if readErr != nil || profile.Name != "Existing author" || profile.HomeAnimation == nil || *profile.HomeAnimation || profile.HomeVerse != "山色入窗" {
					t.Fatalf("profile: %+v, error: %v", profile, readErr)
				}
			} else {
				var menu struct{ Path, PageType string }
				if err = db.Table("sys_menu").Where("menu_code = ?", "existing-page").Take(&menu).Error; err != nil {
					t.Fatal(err)
				}
				if menu.Path != "/journal" || menu.PageType != "internal" {
					t.Fatalf("existing menu: %+v", menu)
				}
			}
		})
	}
}
