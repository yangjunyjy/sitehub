package main

import (
	"path/filepath"
	"strings"
	"testing"

	"sitehub/common/runtimeconf"

	"gorm.io/gorm"
)

func catalogUpgradeDB(t *testing.T, service string) *gorm.DB {
	t.Helper()
	path := filepath.Join(t.TempDir(), service+".db")
	t.Setenv(strings.ToUpper(service)+"_DB_DSN", path)
	if err := run(service, "sqlite", false); err != nil {
		t.Fatal(err)
	}
	db, err := (runtimeconf.Database{Driver: "sqlite", DSN: path, MaxOpen: 1}).Open()
	if err != nil {
		t.Fatal(err)
	}
	pool, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { pool.Close() })
	return db
}

func TestProductV1CatalogUpgrade(t *testing.T) {
	db := catalogUpgradeDB(t, "product")
	for _, query := range []string{
		"UPDATE microservice_schema SET version = 1",
		"DROP TABLE product_release_groups", "DROP TABLE product_release_files", "DROP TABLE product_release_settings",
		"DROP TABLE product_release_install_options", "DROP TABLE product_release_requirements",
		"ALTER TABLE product_documents DROP COLUMN published_revision",
		"ALTER TABLE product_document_revisions DROP COLUMN group_id",
		"ALTER TABLE product_document_revisions DROP COLUMN sort_order",
		"INSERT INTO product_doc_groups (id, version_id, title, revision) VALUES (1, 'v1', 'Start', 1)",
		"INSERT INTO product_documents (id, version_id, group_id, doc_key, slug, title, content_markdown, revision) VALUES (7, 'v1', 1, 'start', 'start', 'Start', 'Existing text', 3)",
		"INSERT INTO product_document_revisions (document_id, revision, content_markdown) VALUES (7, 3, 'Existing text')",
		"INSERT INTO system_configs (config_key, value) VALUES ('website.downloads', 'old catalog')",
	} {
		if err := db.Exec(query).Error; err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 2; i++ {
		if err := run("product", "sqlite", false); err != nil {
			t.Fatal(err)
		}
	}
	var counts struct{ Drafts, History, Files, Settings, Legacy int }
	query := `SELECT
		(SELECT COUNT(*) FROM product_documents WHERE id=7 AND revision=3 AND published_revision=0 AND content_markdown='Existing text') AS drafts,
		(SELECT COUNT(*) FROM product_document_revisions WHERE document_id=7 AND content_markdown='Existing text') AS history,
		(SELECT COUNT(*) FROM product_release_files) AS files,
		(SELECT COUNT(*) FROM product_release_settings) AS settings,
		(SELECT COUNT(*) FROM system_configs WHERE config_key='website.downloads' AND value='old catalog') AS legacy`
	if err := db.Raw(query).Scan(&counts).Error; err != nil {
		t.Fatal(err)
	}
	if counts.Drafts != 1 || counts.History != 1 || counts.Files != 0 || counts.Settings != 1 || counts.Legacy != 1 {
		t.Fatalf("unexpected state after catalog upgrade: %+v", counts)
	}
}

func TestRetiringProductMenuPreservesCustomChildAccess(t *testing.T) {
	db := catalogUpgradeDB(t, "rbac")
	var parentID uint64
	if err := db.Table("sys_menu").Select("id").Where("component = ? AND system_code = ?", "product/docs/index", "admin").Take(&parentID).Error; err != nil {
		t.Fatal(err)
	}
	insertMenu := "INSERT INTO sys_menu (id, menu_type, menu_name, menu_code, component, parent_id) VALUES (?, 2, ?, ?, ?, ?)"
	if err := db.Exec(insertMenu, 10001, "分组", "builtin:page-product-docs-groups-index", "product/docs/groups/index", parentID).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(insertMenu, 10002, "自定义", "custom-doc-child", "custom/index", 10001).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("INSERT INTO sys_role_menu (role_id, menu_id, created_at) VALUES (71, 10001, CURRENT_TIMESTAMP), (71, 10002, CURRENT_TIMESTAMP)").Error; err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if err := run("rbac", "sqlite", false); err != nil {
			t.Fatal(err)
		}
	}
	var counts struct{ Retired, Promoted, Grants, ChildGrants int }
	query := `SELECT
		(SELECT COUNT(*) FROM sys_menu WHERE id=10001 AND deleted=1) AS retired,
		(SELECT COUNT(*) FROM sys_menu WHERE id=10002 AND deleted=0 AND parent_id=?) AS promoted,
		(SELECT COUNT(*) FROM sys_role_menu WHERE role_id=71) AS grants,
		(SELECT COUNT(*) FROM sys_role_menu WHERE role_id=71 AND menu_id=10002) AS child_grants`
	if err := db.Raw(query, parentID).Scan(&counts).Error; err != nil {
		t.Fatal(err)
	}
	if counts.Retired != 1 || counts.Promoted != 1 || counts.Grants != 1 || counts.ChildGrants != 1 {
		t.Fatalf("unexpected menus or grants after retirement: %+v", counts)
	}
}
