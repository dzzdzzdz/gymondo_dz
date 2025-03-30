package database

import (
	"fmt"
	"gymondo_dz/pkg/models"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewPostgresConnection() (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("Connected to PostgreSQL database")
	return db, nil
}

func AutoMigrate(db *gorm.DB, isTest bool) error {
	if isTest {
		// SQLite-specific schema
		err := db.Exec(`
            CREATE TABLE IF NOT EXISTS products (
                id TEXT PRIMARY KEY,
                name TEXT NOT NULL,
                description TEXT,
                price REAL NOT NULL,
                duration INTEGER NOT NULL,
                created_at DATETIME,
                updated_at DATETIME,
                deleted_at DATETIME
            )
        `).Error
		if err != nil {
			return fmt.Errorf("failed to create products table: %w", err)
		}

		err = db.Exec(`
    CREATE TABLE IF NOT EXISTS subscriptions (
        id TEXT PRIMARY KEY,
        user_id TEXT NOT NULL,
        product_id TEXT NOT NULL,
        start_date DATETIME NOT NULL,
        end_date DATETIME NOT NULL,
        status TEXT NOT NULL DEFAULT 'active',
        paused_at DATETIME,
        cancelled_at DATETIME,
        created_at DATETIME,
        updated_at DATETIME,
        deleted_at DATETIME,
        FOREIGN KEY (product_id) REFERENCES products(id)
    )
`).Error

		return err
	}

	// PostgreSQL migrations
	return db.AutoMigrate(&models.Product{}, &models.Subscription{})
}
