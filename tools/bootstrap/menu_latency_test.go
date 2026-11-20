package main

import (
	"path/filepath"
	"testing"
	"time"

	"gorm.io/gorm"
	"sitehub/common/runtimeconf"
	rbacmysql "sitehub/service/rbac/model/mysql"
)

func TestWideMenuReadCost(t *testing.T) {
	db, err := (runtimeconf.Database{Driver: "sqlite", DSN: filepath.Join(t.TempDir(), "menus.db"), MaxOpen: 1}).Open()
	if err != nil {
		t.Fatal(err)
	}
	pool, _ := db.DB()
	defer pool.Close()
	for _, statement := range []string{
		"CREATE TABLE sys_role(id INTEGER PRIMARY KEY,status INTEGER,deleted INTEGER)",
		"CREATE TABLE sys_user_role(user_id INTEGER,role_id INTEGER)",
		"CREATE TABLE sys_role_menu(role_id INTEGER,menu_id INTEGER)",
		"CREATE TABLE sys_menu(id INTEGER PRIMARY KEY,parent_id INTEGER,status INTEGER,deleted INTEGER,sort INTEGER)",
		"INSERT INTO sys_role VALUES (1,1,0),(2,0,0)", "INSERT INTO sys_user_role VALUES (1,1),(1,2)",
		"INSERT INTO sys_menu VALUES (1,0,1,0,0),(200,0,1,0,0)", "INSERT INTO sys_role_menu VALUES (2,200)",
	} {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatal(err)
		}
	}
	for id := 2; id <= 101; id++ {
		if err := db.Exec("INSERT INTO sys_menu VALUES (?,1,1,0,?)", id, 102-id).Error; err != nil {
			t.Fatal(err)
		}
		if err := db.Exec("INSERT INTO sys_role_menu VALUES (1,?)", id).Error; err != nil {
			t.Fatal(err)
		}
	}
	reads := 0
	if err := db.Callback().Query().Before("gorm:query").Register("test:count", func(query *gorm.DB) {
		if !query.DryRun {
			reads++
		}
	}); err != nil {
		t.Fatal(err)
	}
	start := time.Now()
	tree, err := rbacmysql.NewSysMenuRepositoryImpl(db).GetUserMenus(1)
	if err != nil || len(tree) != 1 || tree[0].ID != 1 || len(tree[0].Children) != 100 || tree[0].Children[0].ID != 101 {
		t.Fatalf("incorrect menu tree: %v", err)
	}
	t.Logf("100 granted menus with one shared parent: %d database reads in %s", reads, time.Since(start))
	if reads > 2 {
		t.Errorf("expected at most two reads, got %d", reads)
	}
}
