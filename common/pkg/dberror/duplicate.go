package dberror

import (
	"errors"

	"github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

// IsDuplicate uses driver codes, never SQL text that may contain user input.
func IsDuplicate(err error) bool {
	var mysqlErr *mysql.MySQLError
	var postgresErr *pgconn.PgError
	return errors.Is(err, gorm.ErrDuplicatedKey) ||
		(errors.As(err, &mysqlErr) && mysqlErr.Number == 1062) ||
		(errors.As(err, &postgresErr) && postgresErr.Code == "23505")
}
