package main

import (
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func adminSQLDatabase(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
	connection, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	db, err := gorm.Open(mysql.New(mysql.Config{Conn: connection, SkipInitializeWithVersion: true}), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
		_ = connection.Close()
	})
	t.Setenv("BOOTSTRAP_ADMIN_USERNAME", "initial-admin")
	t.Setenv("BOOTSTRAP_ADMIN_PASSWORD", "a-new-password-that-must-not-replace-the-old-one")
	t.Setenv("BOOTSTRAP_ADMIN_EMAIL", "admin@example.test")
	return db, mock
}

func existingAdminRows(email string, admin, status, deleted int) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "username", "email", "password", "is_admin", "status", "deleted"}).
		AddRow(7, "initial-admin", email, "existing-bcrypt-hash-kept-unchanged", admin, status, deleted)
}

func TestEnsureAdminPreservesExistingCredentials(t *testing.T) {
	db, mock := adminSQLDatabase(t)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `sys_user` WHERE username = ? LIMIT ?")).
		WithArgs("initial-admin", 1).WillReturnRows(existingAdminRows("admin@example.test", 1, 1, 0))
	mock.ExpectQuery("SELECT count\\(\\*\\) FROM `sys_role` JOIN sys_user_role").
		WithArgs(uint64(7), uint64(1), "super_admin", 1, 0).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	// No INSERT or UPDATE is permitted. A different password in the environment
	// must not reset the current administrator or grant new privileges.
	mock.ExpectCommit()
	if err := seedAdminWithOptions(db, true); err != nil {
		t.Fatal(err)
	}
}

func TestEnsureAdminCreatesInitialAccountAndSuperAdminBinding(t *testing.T) {
	db, mock := adminSQLDatabase(t)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT .* FROM `sys_user`").WithArgs("initial-admin", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery("SELECT .* FROM `sys_role`").WithArgs(uint64(1), 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectExec("INSERT INTO `sys_role`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO `sys_user`").WillReturnResult(sqlmock.NewResult(7, 1))
	mock.ExpectExec("INSERT INTO `sys_user_role`").WithArgs(uint64(7), uint64(1)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	if err := seedAdminWithOptions(db, true); err != nil {
		t.Fatal(err)
	}
}

func TestEnsureAdminRejectsExistingAccountMismatch(t *testing.T) {
	for _, sample := range []struct {
		name    string
		email   string
		admin   int
		status  int
		deleted int
	}{
		{"different email", "other@example.test", 1, 1, 0},
		{"ordinary account", "admin@example.test", 0, 1, 0},
		{"disabled account", "admin@example.test", 1, 0, 0},
		{"deleted account", "admin@example.test", 1, 1, 1},
	} {
		t.Run(sample.name, func(t *testing.T) {
			db, mock := adminSQLDatabase(t)
			mock.ExpectBegin()
			mock.ExpectQuery("SELECT .* FROM `sys_user`").WithArgs("initial-admin", 1).
				WillReturnRows(existingAdminRows(sample.email, sample.admin, sample.status, sample.deleted))
			mock.ExpectRollback()
			if err := seedAdminWithOptions(db, true); err == nil || !strings.Contains(err.Error(), "no credentials or permissions changed") {
				t.Fatalf("unexpected result: %v", err)
			}
		})
	}
}

func TestEnsureAdminRequiresExistingSuperAdminRole(t *testing.T) {
	db, mock := adminSQLDatabase(t)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT .* FROM `sys_user`").WithArgs("initial-admin", 1).
		WillReturnRows(existingAdminRows("admin@example.test", 1, 1, 0))
	mock.ExpectQuery("SELECT count\\(\\*\\) FROM `sys_role` JOIN sys_user_role").
		WithArgs(uint64(7), uint64(1), "super_admin", 1, 0).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectRollback()
	if err := seedAdminWithOptions(db, true); err == nil || !strings.Contains(err.Error(), "no active super_admin role") {
		t.Fatalf("unexpected result: %v", err)
	}
}

func TestCreateAdminStillRejectsExistingAccount(t *testing.T) {
	db, mock := adminSQLDatabase(t)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT .* FROM `sys_user`").WithArgs("initial-admin", 1).
		WillReturnRows(existingAdminRows("admin@example.test", 1, 1, 0))
	mock.ExpectRollback()
	if err := seedAdmin(db); err == nil || !strings.Contains(err.Error(), "administrator exists") {
		t.Fatalf("unexpected result: %v", err)
	}
}

func TestEnsureAdminRejectsReusedSuperAdminRoleID(t *testing.T) {
	db, mock := adminSQLDatabase(t)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT .* FROM `sys_user`").WithArgs("initial-admin", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery("SELECT .* FROM `sys_role`").WithArgs(uint64(1), 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "role_code", "status", "deleted"}).AddRow(1, "ordinary-role", 1, 0))
	mock.ExpectRollback()
	if err := seedAdminWithOptions(db, true); err == nil || !strings.Contains(err.Error(), "not the expected active super_admin role") {
		t.Fatalf("unexpected result: %v", err)
	}
}

func TestBootstrapAllPreflightsInputsBeforeOpeningAnyDatabase(t *testing.T) {
	for _, sample := range []struct {
		name        string
		domain      string
		all         bool
		seed        bool
		ensureAdmin bool
		want        string
	}{
		{"mutually exclusive domains", "rbac", true, false, false, "either -all or -service"},
		{"mutually exclusive admin modes", "", true, true, true, "either -admin or -ensure-admin"},
		{"all DSNs required", "", true, false, false, "set PRODUCT_DB_DSN"},
		{"admin belongs to rbac", "blog", false, false, true, "administrator belongs to rbac"},
	} {
		t.Run(sample.name, func(t *testing.T) {
			for _, domain := range []string{"rbac", "file", "blog"} {
				t.Setenv(strings.ToUpper(domain)+"_DB_DSN", "invalid-dsn-must-not-be-opened")
			}
			t.Setenv("PRODUCT_DB_DSN", "")
			if err := runBootstrap(sample.domain, "mysql", sample.all, sample.seed, sample.ensureAdmin); err == nil || !strings.Contains(err.Error(), sample.want) {
				t.Fatalf("unexpected result: %v", err)
			}
		})
	}
	for _, domain := range []string{"rbac", "file", "blog", "product"} {
		t.Setenv(strings.ToUpper(domain)+"_DB_DSN", "invalid-dsn-must-not-be-opened")
	}
	t.Setenv("BOOTSTRAP_ADMIN_USERNAME", "admin")
	t.Setenv("BOOTSTRAP_ADMIN_PASSWORD", "short")
	t.Setenv("BOOTSTRAP_ADMIN_EMAIL", "admin@example.test")
	if err := runBootstrap("", "mysql", true, false, true); err == nil || !strings.Contains(err.Error(), "12–72 byte password") {
		t.Fatalf("administrator was not validated before opening the first database: %v", err)
	}
}
