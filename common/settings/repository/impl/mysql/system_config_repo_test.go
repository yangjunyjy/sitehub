package mysql

import (
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestConfigSnapshotCommitUsesLockedRowAndPreservesFailures(t *testing.T) {
	for _, outcome := range []string{"success", "case-sensitive-conflict", "deleted", "read-error", "write-error", "commit-error"} {
		t.Run(outcome, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			orm, err := gorm.Open(mysql.New(mysql.Config{Conn: db, SkipInitializeWithVersion: true}), &gorm.Config{})
			if err != nil {
				t.Fatal(err)
			}
			cause := errors.New("database failure")
			mock.ExpectBegin()
			read := mock.ExpectQuery("SELECT .* FROM `system_configs` .*FOR UPDATE").WithArgs("website.downloads", 1)
			rows := sqlmock.NewRows([]string{"id", "config_key", "value"})
			switch outcome {
			case "read-error":
				read.WillReturnError(cause)
				mock.ExpectRollback()
			case "deleted":
				read.WillReturnRows(rows)
				mock.ExpectCommit()
			case "case-sensitive-conflict":
				read.WillReturnRows(rows.AddRow(7, "website.downloads", `{"version":"V1"}`))
				mock.ExpectCommit()
			default:
				read.WillReturnRows(rows.AddRow(7, "website.downloads", `{"version":"v1"}`))
				write := mock.ExpectExec(regexp.QuoteMeta("UPDATE `system_configs` SET `value`=?,`updated_at`=? WHERE `id` = ?")).WithArgs(`{"version":"v2"}`, sqlmock.AnyArg(), uint64(7))
				if outcome == "write-error" {
					write.WillReturnError(cause)
					mock.ExpectRollback()
				} else {
					write.WillReturnResult(sqlmock.NewResult(0, 1))
					commit := mock.ExpectCommit()
					if outcome == "commit-error" {
						commit.WillReturnError(cause)
					}
				}
			}
			matched, err := NewSystemConfigRepository(orm).CompareAndSwapValue("website.downloads", `{"version":"v1"}`, `{"version":"v2"}`)
			if outcome == "read-error" || outcome == "write-error" || outcome == "commit-error" {
				if matched || !errors.Is(err, cause) {
					t.Fatalf("failure lost: matched=%v err=%v", matched, err)
				}
			} else if err != nil || matched != (outcome == "success") {
				t.Fatalf("unexpected result: matched=%v err=%v", matched, err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
