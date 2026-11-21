package schema

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestCatalogReferencesActualAdminPagesAndSeparatesPublicMenus(t *testing.T) {
	_, source, _, _ := runtime.Caller(0)
	web := filepath.Join(filepath.Dir(source), "..", "..", "..", "..", "web")
	codes, paths := map[string]bool{}, map[string]bool{}
	roots := map[string]bool{}
	for _, seed := range catalogForTest(t) {
		if codes[seed.Code] {
			t.Fatalf("duplicate stable code %s", seed.Code)
		}
		codes[seed.Code] = true
		if seed.Parent != "" && !codes[seed.Parent] {
			t.Fatalf("parent %s must precede child %s", seed.Parent, seed.Code)
		}
		if seed.Type != 3 {
			key := seed.System + ":" + seed.Path
			if paths[key] {
				t.Fatalf("duplicate route %s", key)
			}
			paths[key] = true
		}
		if seed.Component != "" {
			views := filepath.Join(web, seed.System, "src", "views")
			if _, err := os.Stat(filepath.Join(views, filepath.FromSlash(seed.Component)+".vue")); err != nil {
				t.Fatalf("%s menu points to missing page %s: %v", seed.System, seed.Component, err)
			}
		}
		if seed.System != "admin" && (seed.Permission != "" || seed.Type == 3) {
			t.Fatal("public navigation contains permission grant")
		}
		if seed.Type == 1 {
			roots[seed.Name] = true
		}
	}
	for _, name := range []string{"系统管理", "博客系统", "产品系统", "后台管理系统"} {
		if !roots[name] {
			t.Fatalf("missing root %s", name)
		}
	}
}

func TestCatalogSeedingIsIdempotentAndDoesNotOverwriteCustomizations(t *testing.T) {
	connection, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	db, err := gorm.Open(mysql.New(mysql.Config{Conn: connection, SkipInitializeWithVersion: true}), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	for pass := 0; pass < 2; pass++ {
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE `sys_menu` SET").WillReturnResult(sqlmock.NewResult(0, 0))
		for index, code := range []string{"admin", "product", "blog"} {
			rows := sqlmock.NewRows([]string{"id", "code", "name"})
			if pass == 1 {
				rows.AddRow(index+1, code, "Customized system")
			}
			mock.ExpectQuery("SELECT .*sys_system").WillReturnRows(rows)
			if pass == 0 {
				mock.ExpectExec("INSERT INTO `sys_system`").WillReturnResult(sqlmock.NewResult(int64(index+1), 1))
			}
		}
		for _, column := range []string{"component", "menu_code", "component"} {
			mock.ExpectExec("UPDATE `sys_menu` SET `" + column + "`=").WillReturnResult(sqlmock.NewResult(0, 0))
		}
		for index, seed := range catalogForTest(t) {
			rows := sqlmock.NewRows([]string{"id", "menu_code", "menu_name", "path"})
			if pass == 1 {
				rows.AddRow(index+1, "builtin:"+seed.Code, "Customized title", fmt.Sprintf("/custom/%d", index))
			}
			mock.ExpectQuery("SELECT .*sys_menu").WillReturnRows(rows)
			if pass == 0 {
				mock.ExpectExec("INSERT INTO `sys_menu`").WillReturnResult(sqlmock.NewResult(int64(index+1), 1))
				if seed.Hidden {
					mock.ExpectExec("UPDATE `sys_menu` SET").WillReturnResult(sqlmock.NewResult(0, 1))
				}
			}
		}
		for _, page := range []string{"product/docs/groups", "product/docs/modes", "product/docs/settings", "product/docs/versions", "product/download-settings"} {
			component := page + "/index"
			mock.ExpectQuery("SELECT .*FROM `sys_menu` WHERE menu_code = .*component = .*").
				WithArgs("builtin:page-"+strings.ReplaceAll(component, "/", "-"), component, "admin", 1).
				WillReturnRows(sqlmock.NewRows([]string{"id"}))
		}
		mock.ExpectCommit()
		if err := SeedCatalog(db); err != nil {
			t.Fatal(err)
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func catalogForTest(t *testing.T) []MenuSeed {
	t.Helper()
	seeds, err := MenuCatalog()
	if err != nil {
		t.Fatal(err)
	}
	return seeds
}

func TestDiscoverNewDirectoriesAndPagesWithoutCatalogEntries(t *testing.T) {
	files := fstest.MapFS{
		"guides/intro/index.vue":                 {},
		"guides/intro/details/index.vue":         {},
		"guides/intro/components/card/index.vue": {},
		"guides/private/menu.json":               {Data: []byte("{\"menu\":false}")},
		"guides/private/index.vue":               {},
	}
	seeds, err := viewChildren(files, "blog", ".", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(seeds) != 3 || seeds[0].Type != 1 || seeds[1].Parent != seeds[0].Code ||
		seeds[2].Parent != seeds[1].Code || seeds[2].Component != "guides/intro/details/index" {
		t.Fatalf("unexpected directory tree: %+v", seeds)
	}
}
