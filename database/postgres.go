package database

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

// Open opens a new database connection.
func Open(db *DB) (err error) {
	if db.ConnectionInfo == "" {
		return fmt.Errorf("connectionInfo required")
	}
	db.Gorm, err = gorm.Open(postgres.Open(db.ConnectionInfo), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("err opening gorm postgres connection: %w", err)
	}
	//err = db.Gorm.Migrator().DropTable(domain.User{})
	//if err != nil {
	//	return err
	//}
	if err := db.Gorm.AutoMigrate(domain.User{}, domain.Tweet{}, domain.Follow{}, domain.Like{}); err != nil {
		return fmt.Errorf("err migrating: %w", err)
	}
	return nil
}
