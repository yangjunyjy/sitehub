// Bootstrap initializes an explicitly selected new service database. It never
// upgrades the original application's schema or mixes business domains.
package main

import (
	"errors"
	"flag"
	"fmt"
	"net/mail"
	"os"
	"sitehub/common/runtimeconf"
	blogschema "sitehub/service/blog/model/schema"
	productschema "sitehub/service/product/model/schema"
	entity "sitehub/service/rbac/model/entity"
	rbacschema "sitehub/service/rbac/model/schema"
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
	all := flag.Bool("all", false, "initialize rbac, file, blog and product databases")
	driver := flag.String("driver", "mysql", "mysql or sqlite")
	seed := flag.Bool("admin", false, "create initial rbac administrator from BOOTSTRAP_ADMIN_* environment")
	ensureAdmin := flag.Bool("ensure-admin", false, "create the initial administrator or verify an existing one without changing credentials")
	flag.Parse()
	if err := runBootstrap(*domain, *driver, *all, *seed, *ensureAdmin); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// Validate the complete request before creating tables in the first database.
// Each database still owns its schema marker; a failed run can be repeated.
func runBootstrap(domain, driver string, all, seed, ensureAdmin bool) error {
	if seed && ensureAdmin {
		return fmt.Errorf("select either -admin or -ensure-admin")
	}
	if all && domain != "" {
		return fmt.Errorf("select either -all or -service")
	}
	domains := []string{domain}
	if all {
		domains = []string{"rbac", "file", "blog", "product"}
	} else if schemaFor(domain) == nil {
		return fmt.Errorf("select -all or a valid -service")
	}
	if driver != "mysql" && driver != "sqlite" {
		return fmt.Errorf("select mysql or sqlite driver")
	}
	if (seed || ensureAdmin) && !all && domain != "rbac" {
		return fmt.Errorf("administrator belongs to rbac")
	}
	for _, selected := range domains {
		if os.Getenv(strings.ToUpper(selected)+"_DB_DSN") == "" {
			return fmt.Errorf("set %s_DB_DSN", strings.ToUpper(selected))
		}
	}
	if seed || ensureAdmin {
		if _, err := loadAdminCredentials(); err != nil {
			return err
		}
	}
	for _, selected := range domains {
		if err := runService(selected, driver, seed && selected == "rbac", ensureAdmin && selected == "rbac"); err != nil {
			return fmt.Errorf("initialize %s: %w", selected, err)
		}
	}
	return nil
}

func schemaFor(domain string) func(*gorm.DB) error {
	return map[string]func(*gorm.DB) error{"rbac": rbacschema.Apply, "blog": blogschema.Apply, "product": productschema.Apply, "file": blogschema.ApplyStorage}[domain]
}

func run(domain, driver string, seed bool) error {
	return runService(domain, driver, seed, false)
}

func runService(domain, driver string, seed, ensureAdmin bool) error {
	apply := schemaFor(domain)
	if apply == nil {
		return fmt.Errorf("select a valid service")
	}
	if (seed || ensureAdmin) && domain != "rbac" {
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
	targetVersion := runtimeconf.SchemaVersion(domain)
	if owner.Version < targetVersion {
		if err = apply(db); err != nil {
			return err
		}
		if err = db.Model(&owner).Update("version", targetVersion).Error; err != nil {
			return err
		}
	}
	if domain == "rbac" {
		if err = rbacschema.SeedCatalog(db); err != nil {
			return err
		}
	}
	if seed || ensureAdmin {
		if err = seedAdminWithOptions(db, ensureAdmin); err != nil {
			return err
		}
	}
	if domain == "rbac" {
		if err = rbacschema.AssociateAdminMenus(db); err != nil {
			return err
		}
	}
	fmt.Printf("%s schema is ready\n", domain)
	return nil
}

type adminCredentials struct{ name, password, email string }

func loadAdminCredentials() (adminCredentials, error) {
	credentials := adminCredentials{os.Getenv("BOOTSTRAP_ADMIN_USERNAME"), os.Getenv("BOOTSTRAP_ADMIN_PASSWORD"), os.Getenv("BOOTSTRAP_ADMIN_EMAIL")}
	if len(credentials.name) < 1 || len(credentials.name) > 50 || len(credentials.password) < 12 || len(credentials.password) > 72 {
		return adminCredentials{}, fmt.Errorf("administrator requires username and a 12–72 byte password")
	}
	address, err := mail.ParseAddress(credentials.email)
	if err != nil || address.Address != credentials.email {
		return adminCredentials{}, fmt.Errorf("administrator requires a valid email")
	}
	return credentials, nil
}

func seedAdmin(db *gorm.DB) error { return seedAdminWithOptions(db, false) }

func seedAdminWithOptions(db *gorm.DB, ensure bool) error {
	credentials, err := loadAdminCredentials()
	if err != nil {
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		var existing entity.SysUser
		err := tx.Where("username = ?", credentials.name).Take(&existing).Error
		if err == nil {
			if !ensure {
				return fmt.Errorf("administrator exists; no credentials changed")
			}
			if existing.Username != credentials.name || existing.Email != credentials.email || existing.IsAdmin != 1 || existing.Status != 1 || existing.Deleted != 0 {
				return fmt.Errorf("existing account is not the configured active administrator; no credentials or permissions changed")
			}
			var count int64
			if err = tx.Model(&entity.SysRole{}).
				Joins("JOIN sys_user_role ON sys_role.id = sys_user_role.role_id").
				Where("sys_user_role.user_id = ? AND sys_role.id = ? AND sys_role.role_code = ? AND sys_role.status = ? AND sys_role.deleted = ?", existing.ID, entity.SuperAdminRoleID, "super_admin", 1, 0).
				Count(&count).Error; err != nil {
				return err
			}
			if count != 1 {
				return fmt.Errorf("existing administrator has no active super_admin role; no credentials or permissions changed")
			}
			return nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		var role entity.SysRole
		err = tx.First(&role, entity.SuperAdminRoleID).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			role = entity.SysRole{RoleName: "超级管理员", RoleCode: "super_admin", Status: 1}
			role.ID = entity.SuperAdminRoleID
			if err = tx.Create(&role).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else if role.RoleCode != "super_admin" || role.Status != 1 || role.Deleted != 0 {
			return fmt.Errorf("initial administrator role is not the expected active super_admin role; no account created")
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(credentials.password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		user := entity.SysUser{Username: credentials.name, Nickname: credentials.name, Email: credentials.email, Password: string(hash), Status: 1, IsAdmin: 1}
		if err := tx.Create(&user).Error; err != nil {
			return err
		}
		return tx.Create(&entity.SysUserRole{UserID: user.ID, RoleID: entity.SuperAdminRoleID}).Error
	})
}
