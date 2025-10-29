package database

import (
	"time"

	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/lib/pq"

	"github.com/newt239/chat/ent"
)

// NewConnection creates a new ent client connection
func NewConnection(dsn string) (*ent.Client, error) {
	drv, err := entsql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	client := ent.NewClient(ent.Driver(drv))
	return client, nil
}

// InitDB initializes the database connection with default settings
func InitDB(dsn string) (*ent.Client, error) {
	var client *ent.Client
	var err error

	// Retry connection with exponential backoff
	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		client, err = NewConnection(dsn)
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
	// Note: ent client doesn't expose DB directly, connection pool is managed by the driver

	return client, nil
}
