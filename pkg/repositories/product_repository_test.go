package repositories_test

import (
	"gymondo_dz/pkg/database"
	"gymondo_dz/pkg/models"
	"gymondo_dz/pkg/repositories"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type ProductRepositoryTestSuite struct {
	suite.Suite
	db   *gorm.DB
	repo repositories.ProductRepository
}

func (s *ProductRepositoryTestSuite) SetupTest() {
	// Setup in-memory SQLite database with foreign keys enabled
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		s.FailNow("Failed to connect to test database")
	}

	// Enable foreign key constraints for SQLite
	db.Exec("PRAGMA foreign_keys = ON")

	// Run migrations
	if err := database.AutoMigrate(db, true); err != nil {
		s.FailNow("Failed to migrate test database: " + err.Error())
	}

	s.db = db
	s.repo = repositories.NewProductRepository(db)
}

func (s *ProductRepositoryTestSuite) BeforeTest(suiteName, testName string) {
	// Clear all data before each test
	s.db.Exec("DELETE FROM products")
	s.db.Exec("DELETE FROM subscriptions")

	// Seed fresh test data
	testProducts := []models.Product{
		{
			ID:          uuid.MustParse("11111111-1111-1111-1111-111111111111"),
			Name:        "Monthly Plan",
			Description: "1 month subscription",
			Price:       9.99,
			Duration:    models.DurationMonth,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			Name:        "Yearly Plan",
			Description: "1 year subscription",
			Price:       99.99,
			Duration:    models.DurationYear,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.MustParse("33333333-3333-3333-3333-333333333333"),
			Name:        "Lifetime Plan",
			Description: "Lifetime access",
			Price:       999.99,
			Duration:    models.DurationLifetime,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	// Create products one by one to better handle errors
	for _, p := range testProducts {
		if err := s.db.Create(&p).Error; err != nil {
			s.FailNow("Failed to seed test data: " + err.Error())
		}
	}
}

func TestProductRepositorySuite(t *testing.T) {
	suite.Run(t, new(ProductRepositoryTestSuite))
}

func (s *ProductRepositoryTestSuite) TestGetProducts() {
	products, err := s.repo.GetProducts()

	s.NoError(err)
	s.Len(products, 3)

	// Verify first product
	s.Equal("Monthly Plan", products[0].Name)
	s.Equal(models.DurationMonth, products[0].Duration)
	s.Equal(9.99, products[0].Price)
}

func (s *ProductRepositoryTestSuite) TestGetProduct() {
	tests := []struct {
		name          string
		id            string
		expectError   bool
		expectedError string
		validate      func(*models.Product)
	}{
		{
			name:        "Valid existing product",
			id:          "11111111-1111-1111-1111-111111111111",
			expectError: false,
			validate: func(p *models.Product) {
				s.Equal("Monthly Plan", p.Name)
				s.Equal(9.99, p.Price)
			},
		},
		{
			name:          "Non-existent product",
			id:            "00000000-0000-0000-0000-000000000000",
			expectError:   true,
			expectedError: "product not found",
		},
		{
			name:          "Malformed UUID",
			id:            "not-a-uuid",
			expectError:   true,
			expectedError: "invalid product ID format",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			product, err := s.repo.GetProduct(tt.id)

			if tt.expectError {
				s.Error(err)
				s.Equal(tt.expectedError, err.Error())
				s.Nil(product)
			} else {
				s.NoError(err)
				s.NotNil(product)
				if tt.validate != nil {
					tt.validate(product)
				}
			}
		})
	}
}
