package database

import (
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewConnection(dsn string, logLevel logger.LogLevel) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, err
	}

	return db, nil
}

// InitDB initializes the database connection with default settings
func InitDB(dsn string) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	// Retry connection with exponential backoff
	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		db, err = NewConnection(dsn, logger.Info)
		if err == nil {
			break
		}

		if i < maxRetries-1 {
			time.Sleep(time.Duration(i+1) * time.Second)
		}
	}

	if err != nil {
		return nil, err
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}
