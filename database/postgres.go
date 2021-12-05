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
	Gorm *gorm.DB
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
	db.Gorm, err = gorm.Open(postgres.Open(db.DSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("err opening gorm postgres connection: %w", err)
	}
	err = db.Gorm.Migrator().DropTable(domain.User{}, domain.OAuth{})
	if err != nil {
		return err
	}
	if err := db.Gorm.AutoMigrate(domain.User{}, domain.OAuth{}); err != nil {
		return fmt.Errorf("err migrating: %w", err)
	}
	return nil
}
