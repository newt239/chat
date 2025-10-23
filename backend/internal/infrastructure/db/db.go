package db

import (
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
	db, err := NewConnection(dsn, logger.Info)
	if err != nil {
		return nil, err
	}

	// Run auto migrations
	if err := AutoMigrate(db); err != nil {
		return nil, err
	}

	return db, nil
}

// AutoMigrate runs automatic migrations for all models
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&User{},
		&Session{},
		&Workspace{},
		&WorkspaceMember{},
		&Channel{},
		&ChannelMember{},
		&Message{},
		&MessageReaction{},
		&ChannelReadState{},
		&Attachment{},
	)
}
