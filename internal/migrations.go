package internal

import (
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations() {
	connString := os.Getenv("DATABASE_URL")

	m, err := migrate.New("file://../db/migrations", connString)
	if err != nil {
		log.Fatal("Migration setup failed:", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal("Migration failed:", err)
	}
}
