package main

import (
	//"context"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"wtfTwitter/domain"
)

// DB provides the database connection.
type DB struct {
	// Object-relational mapping.
	Gorm *gorm.DB
	// Connection info string containing database name, user, port etc.
	ConnectionInfo string
}

// NewDB returns a new instance of DB.
func NewDB(connectionInfo string) *DB {
	db := &DB{
		ConnectionInfo: connectionInfo,
	}
	return db
}

// Open opens a new database connection. It also configures logging
// based on whether we're in development or in production.
func Open(db *DB, isProd bool) (err error) {
	if db.ConnectionInfo == "" {
		return fmt.Errorf("connectionInfo required")
	}
	logMode := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}
	if isProd == false {
		logMode.Logger = logger.Default.LogMode(logger.Info)
	}
	db.Gorm, err = gorm.Open(postgres.Open(db.ConnectionInfo), logMode)
	if err != nil {
		return fmt.Errorf("err opening gorm postgres connection: %w", err)
	}
	return nil
}

// AutoMigrate runs database migrations for all tables.
func AutoMigrate(db *DB) error {
	return db.Gorm.AutoMigrate(
		domain.User{},
		domain.OAuth{},
		domain.Tweet{},
		domain.Follow{},
		domain.Like{},
	)
}

// DestructiveReset drops all tables and rebuilds them.
func DestructiveReset(db *DB) error {
	err := db.Gorm.Migrator().DropTable(
		domain.User{},
		domain.OAuth{},
		domain.Tweet{},
		domain.Follow{},
		domain.Like{},
	)
	if err != nil {
		return err
	}
	return AutoMigrate(db)
}

// Close closes the database connection.
func Close (db *DB) error {
	sqlDb, _ := db.Gorm.DB()
	return sqlDb.Close()
}