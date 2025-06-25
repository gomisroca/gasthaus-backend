package internal

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)


func SetupDB() (*pgxpool.Pool, error) {
	// Load environment variables from .env file
	err := godotenv.Load()
    if err != nil {
        fmt.Println("Error loading .env file")
    }	
	
	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		return nil, fmt.Errorf("DATABASE_URL not set in environment")
	}

	// Database connection setup
	dbpool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
        return nil, fmt.Errorf("unable to connect to database: %v", err)
	}

	return dbpool, nil
}