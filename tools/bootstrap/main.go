// Bootstrap initializes an explicitly selected new service database. It never
// upgrades the original application's schema or mixes business domains.
package main

import (
	"flag"
	"fmt"
	"sitehub/common/runtimeconf"
	blogschema "sitehub/service/blog/model/schema"
	fileschema "sitehub/service/file/model/schema"
	productschema "sitehub/service/product/model/schema"
	entity "sitehub/service/rbac/model/entity"
	rbacschema "sitehub/service/rbac/model/schema"
	"net/mail"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type schemaOwner struct {
	ID      uint `gorm:"primaryKey"`
	Service string
	Version int
}

func (schemaOwner) TableName() string { return "microservice_schema" }
func main() {
	domain := flag.String("service", "", "rbac, blog, product or file")
	driver := flag.String("driver", "mysql", "mysql or sqlite")
	seed := flag.Bool("admin", false, "create initial rbac administrator from BOOTSTRAP_ADMIN_* environment")
	flag.Parse()
	if err := run(*domain, *driver, *seed); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run(domain, driver string, seed bool) error {
	apply := map[string]func(*gorm.DB) error{"rbac": rbacschema.Apply, "blog": blogschema.Apply, "product": productschema.Apply, "file": fileschema.Apply}[domain]
	if apply == nil {
		return fmt.Errorf("select a valid service")
	}
	if seed && domain != "rbac" {
		return fmt.Errorf("administrator belongs to rbac")
	}
	dsn := os.Getenv(strings.ToUpper(domain) + "_DB_DSN")
	if dsn == "" {
		return fmt.Errorf("set %s_DB_DSN", strings.ToUpper(domain))
	}
	db, err := (runtimeconf.Database{Driver: driver, DSN: dsn, MaxOpen: 1, MaxIdle: 1}).Open()
	if err != nil {
		return err
	}
	pool, _ := db.DB()
	defer pool.Close()
	tables, err := db.Migrator().GetTables()
	if err != nil {
		return err
	}
	if !db.Migrator().HasTable(&schemaOwner{}) {
		for _, table := range tables {
			if !strings.HasPrefix(table, "sqlite_") {
				return fmt.Errorf("database is not an empty service database; refusing to change %s", table)
			}
		}
		if err = db.AutoMigrate(&schemaOwner{}); err != nil {
			return err
		}
		if err = db.Create(&schemaOwner{ID: 1, Service: domain}).Error; err != nil {
			return err
		}
	}
	var owner schemaOwner
	if err = db.First(&owner, 1).Error; err != nil {
		return err
	}
	if owner.Service != domain {
		return fmt.Errorf("database belongs to %s", owner.Service)
	}
	if owner.Version < 1 {
		if err = apply(db); err != nil {
			return err
		}
		if err = db.Model(&owner).Update("version", 1).Error; err != nil {
			return err
		}
	}
	if seed {
		return seedAdmin(db)
	}
	fmt.Printf("%s schema is ready\n", domain)
	return nil
}
func seedAdmin(db *gorm.DB) error {
	name, password, email := os.Getenv("BOOTSTRAP_ADMIN_USERNAME"), os.Getenv("BOOTSTRAP_ADMIN_PASSWORD"), os.Getenv("BOOTSTRAP_ADMIN_EMAIL")
	if len(name) < 1 || len(name) > 50 || len(password) < 12 || len(password) > 72 {
		return fmt.Errorf("administrator requires username and a 12–72 byte password")
	}
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email {
		return fmt.Errorf("administrator requires a valid email")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&entity.SysUser{}).Where("username = ?", name).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return fmt.Errorf("administrator exists; no credentials changed")
		}
		role := entity.SysRole{}
		role.ID = 1
		role.RoleName = "超级管理员"
		role.RoleCode = "super_admin"
		role.Status = 1
		if err := tx.FirstOrCreate(&role, 1).Error; err != nil {
			return err
		}
		user := entity.SysUser{Username: name, Nickname: name, Email: email, Password: string(hash), Status: 1, IsAdmin: 1}
		if err := tx.Create(&user).Error; err != nil {
			return err
		}
		return tx.Create(&entity.SysUserRole{UserID: user.ID, RoleID: 1}).Error
	})
}
