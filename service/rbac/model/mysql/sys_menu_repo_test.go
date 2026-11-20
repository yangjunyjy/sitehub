package mysql

import (
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	driver "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	rbacrepository "sitehub/service/rbac/model"
	rbacentity "sitehub/service/rbac/model/entity"
)

func menuSQLStore(t *testing.T) (rbacrepository.SysMenuRepository, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	orm, err := gorm.Open(mysql.New(mysql.Config{Conn: db, SkipInitializeWithVersion: true}), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
	})
	return NewSysMenuRepositoryImpl(orm), mock
}

func TestMenuInsertPreservesDisabledAndHiddenValues(t *testing.T) {
	repo, mock := menuSQLStore(t)
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `sys_menu` (`component`,`create_by`,`create_time`,`external_url`,`icon`,`is_visible`,`menu_code`,`menu_name`,`menu_type`,`page_type`,`parent_id`,`path`,`perms`,`remark`,`sort`,`status`,`system_code`,`update_by`,`update_time`) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)")).
		WithArgs("", uint64(7), sqlmock.AnyArg(), "https://example.org", "", false, "menu:unique", "external", 2, "external", 0, "", "", "", 0, 0, "admin", uint64(7), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(19, 1))
	mock.ExpectCommit()
	menu := &rbacentity.SysMenu{BaseModel: rbacentity.BaseModel{CreateBy: 7}, SystemCode: "admin", MenuName: "external", MenuCode: "menu:unique", MenuType: 2, PageType: "external", ExternalURL: "https://example.org", Status: 0, IsVisible: false}
	if err := repo.Create(menu); err != nil {
		t.Fatal(err)
	}
	if menu.ID != 19 || menu.CreateTime == nil {
		t.Fatalf("insert result not returned: %+v", menu)
	}
}

func TestMenuNamesAndPathsUseTheirSystemNamespace(t *testing.T) {
	repo, mock := menuSQLStore(t)
	row := func() *sqlmock.Rows { return sqlmock.NewRows([]string{"id"}).AddRow(5) }
	mock.ExpectQuery("system_code = .* AND parent_id = .* AND menu_name = .* AND deleted = 0").
		WithArgs("blog", 0, "首页", 1).WillReturnRows(row())
	if _, err := repo.FindByName("blog", 0, "首页"); err != nil {
		t.Fatal(err)
	}
	mock.ExpectQuery("system_code = .* AND path = .* AND deleted = 0").
		WithArgs("product", "/", 1).WillReturnRows(row())
	if _, err := repo.FindByPath("product", 9, "/"); err != nil {
		t.Fatal(err)
	}
	mock.ExpectQuery("system_code = .* AND path = .* AND deleted = 0.*parent_id = ").
		WithArgs("admin", "articles", 9, 1).WillReturnRows(row())
	if _, err := repo.FindByPath("admin", 9, "articles"); err != nil {
		t.Fatal(err)
	}
}

func TestMenuLockedUpdatePreservesUnsubmittedColumnsAndRollsBack(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(map[bool]string{false: "commit", true: "rollback"}[fail], func(t *testing.T) {
			repo, mock := menuSQLStore(t)
			cause := &driver.MySQLError{Number: 1062, Message: "private duplicate index"}
			mock.ExpectBegin()
			mock.ExpectQuery("SELECT .* FROM `sys_menu` .*FOR UPDATE").WithArgs(uint64(9), 1).
				WillReturnRows(sqlmock.NewRows([]string{"id", "menu_code", "create_by", "perms"}).AddRow(9, "existing", 7, "existing:permission"))
			// System and permission are editable; creation metadata remains intact.
			write := mock.ExpectExec(regexp.QuoteMeta("UPDATE `sys_menu` SET `component`=?,`external_url`=?,`icon`=?,`is_visible`=?,`menu_code`=?,`menu_name`=?,`menu_type`=?,`page_type`=?,`parent_id`=?,`path`=?,`perms`=?,`remark`=?,`sort`=?,`status`=?,`system_code`=?,`update_by`=?,`update_time`=? WHERE id = ?")).
				WithArgs("", "", "", false, "existing", "renamed", 1, "internal", 0, "", "updated:permission", "", 0, 0, "admin", uint64(8), sqlmock.AnyArg(), uint64(9))
			if fail {
				write.WillReturnError(cause)
				mock.ExpectRollback()
			} else {
				write.WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			}
			err := repo.InTransaction(func(tx rbacrepository.SysMenuRepository) error {
				row, err := tx.FindByIDForUpdate(9)
				if err != nil {
					return err
				}
				return tx.Update(&rbacentity.SysMenu{BaseModel: rbacentity.BaseModel{ID: 9, UpdateBy: 8}, SystemCode: "admin", Permission: "updated:permission", MenuName: "renamed", MenuType: 1, PageType: "internal", MenuCode: row.MenuCode})
			})
			if fail && !errors.Is(err, cause) {
				t.Fatalf("driver cause lost: %v", err)
			}
			if !fail && err != nil {
				t.Fatal(err)
			}
		})
	}
}
