package main

import (
	"sitehub/common/runtimeconf"
	"os"
	"path/filepath"
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
