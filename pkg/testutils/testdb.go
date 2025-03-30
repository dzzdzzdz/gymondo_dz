package testutils

import (
	"fmt"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func SetupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	// Connect to default postgres database to create our test database
	adminDSN := "host=localhost user=postgres password=postgres dbname=postgres port=5432 sslmode=disable TimeZone=UTC"
	adminDB, err := gorm.Open(postgres.Open(adminDSN), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to admin database: %v", err)
	}

	// Create test database if it doesn't exist
	testDBName := "test_gymondo_" + t.Name()
	err = adminDB.Exec(fmt.Sprintf("CREATE DATABASE %s", testDBName)).Error
	if err != nil {
		// Database might already exist, try to continue
		t.Logf("Could not create test database (might already exist): %v", err)
	}

	// Connect to the test database
	testDSN := fmt.Sprintf("host=localhost user=postgres password=postgres dbname=%s port=5432 sslmode=disable TimeZone=UTC", testDBName)
	db, err := gorm.Open(postgres.Open(testDSN), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Clean tables (use TRUNCATE for proper cleanup)
	tables := []string{"subscriptions", "products"}
	for _, table := range tables {
		db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
	}

	return db
}

func TeardownTestDB(t *testing.T, db *gorm.DB) {
	t.Helper()
	if db != nil {
		sqlDB, err := db.DB()
		if err != nil {
			t.Logf("Failed to get SQL DB: %v", err)
			return
		}
		sqlDB.Close()
	}
}
