package internal

import (
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations() {
	// Load the DB connection string from env
	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		log.Fatal("DATABASE_URL environment variable not set")
	}

	// Determine the migration path
	migrationPath := os.Getenv("MIGRATIONS_PATH")
	if migrationPath == "" {
		// Default relative path from the working directory
		migrationPath = "db/migrations"
	}

	// Debug info
	wd, _ := os.Getwd()
	fmt.Printf("Running migrations from '%s' (working directory: %s)\n", migrationPath, wd)

	// Create the migrate instance
	m, err := migrate.New(
		"file://"+migrationPath,
		connString,
	)
	if err != nil {
		log.Fatal("Migration setup failed:", err)
	}

	// Apply migrations
	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			fmt.Println("No new migrations to apply.")
		} else {
			log.Fatal("Migration failed:", err)
		}
	} else {
		fmt.Println("Migrations applied successfully!")
	}
}
