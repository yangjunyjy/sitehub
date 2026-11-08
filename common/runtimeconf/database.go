package runtimeconf

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Database struct {
	Driver  string `json:",default=mysql"`
	DSN     string
	MaxOpen int `json:",default=20"`
	MaxIdle int `json:",default=5"`
}

// Open is called exclusively by a service's dependency construction entry point.
func (c Database) Open() (*gorm.DB, error) {
	var dialector gorm.Dialector
	switch c.Driver {
	case "mysql":
		dialector = mysql.Open(c.DSN)
	case "sqlite":
		dialector = sqlite.Open(c.DSN)
	default:
		return nil, fmt.Errorf("unsupported database driver %q", c.Driver)
	}
	// Queries retain timing/error visibility without exposing passwords, session
	// tokens or storage credentials through bound SQL parameters.
	sqlLog := logger.New(log.New(os.Stderr, "", log.LstdFlags), logger.Config{
		SlowThreshold: time.Second, LogLevel: logger.Warn,
		IgnoreRecordNotFoundError: true, ParameterizedQueries: true,
	})
	db, err := gorm.Open(dialector, &gorm.Config{TranslateError: true, DisableForeignKeyConstraintWhenMigrating: true, Logger: sqlLog})
	if err != nil {
		return nil, err
	}
	pool, err := db.DB()
	if err != nil {
		return nil, err
	}
	pool.SetMaxOpenConns(c.MaxOpen)
	pool.SetMaxIdleConns(c.MaxIdle)
	pool.SetConnMaxLifetime(30 * time.Minute)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err = pool.PingContext(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return db, nil
}
