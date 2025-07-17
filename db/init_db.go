package db

import (
	"context"

	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectDB() *gorm.DB {
	dsn := os.Getenv("DB_URL")
	if dsn == "" {
		log.Fatal("❌ DB_URL env var is required")
	}

	// dbHost := os.Getenv("DB_HOST")
	// dbPort := os.Getenv("DB_PORT")
	// dbUser := os.Getenv("DB_USER")
	// dbPassword := os.Getenv("DB_PASSWORD")
	// dbName := os.Getenv("DB_NAME")

	// log.Printf("Connecting to database at: %s:%s as %s (DB: %s)", dbHost, dbPort, dbUser, dbName)

	// dsn := fmt.Sprintf(
	// 	"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=UTC",
	// 	dbHost, dbPort, dbUser, dbPassword, dbName,
	// )

	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ Failed to connect to database using GORM: %v", err)
	}

	// Ping check
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := gormDB.WithContext(ctx).Raw("SELECT 1").Error; err != nil {
		log.Fatalf("❌ Database ping failed: %v", err)
	}

	log.Println("✅ GORM database connection established successfully!")

	// ✅ Run migrations
	RunMigrations()

	return gormDB
}


