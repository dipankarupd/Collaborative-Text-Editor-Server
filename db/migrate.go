package db

import (
	"log"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // Import postgres database driver
	_ "github.com/golang-migrate/migrate/v4/source/file"       // Import file source driver
)

func RunMigrations() {
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("❌ DB_URL env var not set for migrations")
	}

	// Point to local file path for migrations
	wd, _ := os.Getwd()
	path := "file://" + filepath.Join(wd, "db", "migrations")

	m, err := migrate.New(
		path,
		dbURL,
	)
	if err != nil {
		log.Fatalf("❌ Failed to initialize migrations: %v", err)
	}

	// Run all up migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("❌ Migration failed: %v", err)
	}

	log.Println("✅ Migrations applied successfully")
}