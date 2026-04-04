package internal

import (
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations() error {
	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		return fmt.Errorf("DATABASE_URL environment variable not set")
	}

	migrationPath := os.Getenv("MIGRATIONS_PATH")
	if migrationPath == "" {
		migrationPath = "db/migrations"
	}

	wd, _ := os.Getwd()
	fmt.Printf("Running migrations from '%s' (working directory: %s)\n", migrationPath, wd)

	m, err := migrate.New(
		"file://"+migrationPath,
		connString,
	)
	if err != nil {
		return fmt.Errorf("migration setup failed: %w", err)
	}

	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			fmt.Println("No new migrations to apply.")
			return nil
		}
		return fmt.Errorf("migration failed: %w", err)
	}

	fmt.Println("Migrations applied successfully!")
	return nil
}