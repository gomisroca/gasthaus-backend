package internal

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func SetupDB() (*pgxpool.Pool, error) {
	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		return nil, fmt.Errorf("DATABASE_URL not set in environment")
	}

	// Database connection setup
	dbpool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	if err := dbpool.Ping(context.Background()); err != nil {
		dbpool.Close()
		return nil, fmt.Errorf("database unreachable: %w", err)
	}
	
	return dbpool, nil
}
