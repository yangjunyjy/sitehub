package main

import (
	"path/filepath"
	"sitehub/common/runtimeconf"
	"sitehub/service/rbac/domain"
	"testing"
)

func TestMemberUpgradeFromVersionFour(t *testing.T) {
	dsn := filepath.Join(t.TempDir(), "members.db")
	t.Setenv("RBAC_DB_DSN", dsn)
	db, err := (runtimeconf.Database{Driver: "sqlite", DSN: dsn, MaxOpen: 1}).Open()
	if err != nil {
		t.Fatal(err)
	}
	pool, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	if err = db.AutoMigrate(&schemaOwner{}); err != nil {
		t.Fatal(err)
	}
	if err = db.Create(&schemaOwner{ID: 1, Service: "rbac", Version: 4}).Error; err != nil {
		t.Fatal(err)
	}
	if err = db.Exec("CREATE TABLE product_members (id integer PRIMARY KEY, email text, password_hash text, created_at datetime)").Error; err != nil {
		t.Fatal(err)
	}
	if err = db.Exec("INSERT INTO product_members (id,email) VALUES (1,'one@example.test'),(2,'two@example.test')").Error; err != nil {
		t.Fatal(err)
	}
	if err = run("rbac", "sqlite", false); err != nil {
		t.Fatal(err)
	}
	var users []domain.Member
	if err = db.Select("username,nickname,avatar,email").Find(&users).Error; err != nil || len(users) != 2 || users[0].Username != nil {
		t.Fatalf("old members: %v, %v", users, err)
	}
	var menus int64
	if err = db.Table("sys_menu").Where("component = ?", "blog/about/index").Count(&menus).Error; err != nil || menus != 1 {
		t.Fatalf("about menu: %d, %v", menus, err)
	}
}
