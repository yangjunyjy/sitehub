package services

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"sitehub/service/rbac/model/entity"
)

func systemStore(t *testing.T) (*SysSystemService, sqlmock.Sqlmock) {
	t.Helper()
	connection, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { connection.Close() })
	db, err := gorm.Open(mysql.New(mysql.Config{Conn: connection, SkipInitializeWithVersion: true}), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
	})
	return NewSysSystemService(db), mock
}

func TestMenuPageDestination(t *testing.T) {
	menu := entity.SysMenu{MenuName: "友站", MenuType: 2, PageType: "external", Path: "/old", Component: "old/index", ExternalURL: "https://example.org/notes?q=1"}
	if err := prepareMenu(&menu); err != nil || menu.Path != "" || menu.Component != "" {
		t.Fatalf("external destination: %+v, %v", menu, err)
	}
	for _, address := range []string{"", "//example.org", "javascript:alert(1)", "https://user:pass@example.org", "/relative"} {
		menu.ExternalURL = address
		if err := prepareMenu(&menu); err == nil {
			t.Errorf("accepted %q", address)
		}
	}
	menu.PageType = ""
	if err := prepareMenu(&menu); err != nil || menu.PageType != "internal" || menu.ExternalURL != "" {
		t.Fatalf("legacy destination: %+v, %v", menu, err)
	}
}

func TestSystemOriginAndInternalAdminValidation(t *testing.T) {
	for _, origin := range []string{"javascript:alert(1)", "https://user:secret@example.test", "https://example.test/path", "https://example.test?query=1"} {
		if err := validateSystem(&entity.SysSystem{Code: "product", Name: "Product", Origin: origin}); err == nil {
			t.Fatalf("invalid origin accepted %s", origin)
		}
	}
	if err := validateSystem(&entity.SysSystem{Code: "admin", Name: "Admin", Public: true}); err == nil {
		t.Fatal("internal admin made public")
	}
	value := entity.SysSystem{Code: "product", Name: "Product", Origin: "https://example.test/"}
	if err := validateSystem(&value); err != nil || value.Origin != "https://example.test" {
		t.Fatalf("origin normalization: %s %v", value.Origin, err)
	}
}

func TestSystemCodeCannotChange(t *testing.T) {
	service, mock := systemStore(t)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT .*sys_system.*FOR UPDATE").WillReturnRows(sqlmock.NewRows([]string{"id", "code"}).AddRow(2, "product"))
	mock.ExpectRollback()
	if err := service.Save(context.Background(), &entity.SysSystem{BaseModel: entity.BaseModel{ID: 2}, Code: "renamed", Name: "Product"}); err == nil {
		t.Fatal("immutable system code changed")
	}
}

func TestPublicMenusRespectSystemGateAndQueryOnlyVisibleNonButtons(t *testing.T) {
	for _, code := range []string{"admin", "disabled", "private"} {
		t.Run(code, func(t *testing.T) {
			service, mock := systemStore(t)
			mock.ExpectQuery("SELECT .*sys_system").WillReturnRows(sqlmock.NewRows([]string{"id", "code", "enabled", "public"}).AddRow(1, code, code != "disabled", code != "private"))
			items, err := service.MenuTree(context.Background(), code, true)
			if err != nil || len(items) != 0 {
				t.Fatalf("private system exposed: %v %v", items, err)
			}
		})
	}
	service, mock := systemStore(t)
	mock.ExpectQuery("SELECT .*sys_system").WillReturnRows(sqlmock.NewRows([]string{"id", "code", "enabled", "public"}).AddRow(2, "product", true, true))
	mock.ExpectQuery("SELECT .*sys_menu.*system_code = .*status = .*is_visible = .*menu_type IN").WithArgs("product", 0, 1, true, 1, 2).WillReturnRows(sqlmock.NewRows([]string{"id", "system_code", "menu_name", "path", "menu_type", "status", "is_visible"}).AddRow(9, "product", "Downloads", "/download", 2, 1, true))
	items, err := service.MenuTree(context.Background(), "product", true)
	if err != nil || len(items) != 1 || items[0].SystemCode != "product" {
		t.Fatalf("filtered public navigation: %v %v", items, err)
	}
}

func TestMenuParentSystemBoundary(t *testing.T) {
	service, mock := systemStore(t)
	mock.ExpectQuery("SELECT .*sys_system").WillReturnRows(sqlmock.NewRows([]string{"id", "code"}).AddRow(2, "product"))
	mock.ExpectQuery("SELECT .*sys_menu").WillReturnRows(sqlmock.NewRows([]string{"id", "system_code"}).AddRow(1, "admin"))
	if err := service.ValidateMenu(context.Background(), &entity.SysMenu{ParentID: 1, SystemCode: "product"}); err == nil {
		t.Fatal("cross-system menu parent accepted")
	}
}

func TestPermissionIsIndependentFromMenuCode(t *testing.T) {
	s := &sysPermissionService{}
	menu := &entity.SysMenu{MenuCode: "system:user:delete", Permission: "system:user:list"}
	permissions := s.extractPermissions(menu)
	if len(permissions) != 1 || permissions[0] != "system:user:list" || s.checkMenuPermission([]*entity.SysMenu{menu}, "system:user:delete") {
		t.Fatal("menu identity granted API permission")
	}
}
