package database

import (
	//"context"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"wtfTwitter/domain"
)

type DB struct {
	db *gorm.DB
	//ctx context.Context

	DSN string
}

func NewDB(dsn string) *DB {
	db := &DB{
		DSN: dsn,
	}
	return db
}

func Open(db *DB) (err error) {
	if db.DSN == "" {
		return fmt.Errorf("dsn required")
	}
	db.db, err = gorm.Open(postgres.Open(db.DSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("err opening gorm postgres connection: %w", err)
	}
	if err := db.db.AutoMigrate(domain.User{}, domain.Auth{}); err != nil {
		return fmt.Errorf("err migrating: %w", err)
	}
	return nil
}
