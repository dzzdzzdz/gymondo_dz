package testutils

import (
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

func (s *ProductRepositoryTestSuite) SetupSuite() {
	var err error
	s.db, err = gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		s.FailNow("Failed to connect to test database")
	}

	// Create tables
	err = s.db.AutoMigrate(&models.Product{})
	if err != nil {
		s.FailNow("Failed to migrate database: " + err.Error())
	}

	s.repo = repositories.NewProductRepository(s.db)
}

func (s *ProductRepositoryTestSuite) SetupTest() {
	// Clear and seed data for each test
	s.db.Exec("DELETE FROM products")

	// Create test products
	testProducts := []models.Product{
		{
			ID:          uuid.MustParse("11111111-1111-1111-1111-111111111111"),
			Name:        "Monthly Plan",
			Description: "1 month subscription",
			Price:       9.99,
			Duration:    models.DurationMonth,
			CreatedAt:   time.Now().Add(-3 * time.Hour),
			UpdatedAt:   time.Now().Add(-3 * time.Hour),
		},
		{
			ID:          uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			Name:        "Yearly Plan",
			Description: "1 year subscription",
			Price:       99.99,
			Duration:    models.DurationYear,
			CreatedAt:   time.Now().Add(-2 * time.Hour),
			UpdatedAt:   time.Now().Add(-2 * time.Hour),
		},
		{
			ID:          uuid.MustParse("33333333-3333-3333-3333-333333333333"),
			Name:        "Lifetime Plan",
			Description: "Lifetime access",
			Price:       999.99,
			Duration:    models.DurationLifetime,
			CreatedAt:   time.Now().Add(-1 * time.Hour),
			UpdatedAt:   time.Now().Add(-1 * time.Hour),
		},
	}

	for _, p := range testProducts {
		result := s.db.Create(&p)
		if result.Error != nil {
			s.FailNow("Failed to seed test data: " + result.Error.Error())
		}
	}
}

func TestProductRepositorySuite(t *testing.T) {
	suite.Run(t, new(ProductRepositoryTestSuite))
}

func (s *ProductRepositoryTestSuite) TestGetProducts() {
	products, total, err := s.repo.GetProducts(1, 10)
	s.NoError(err)
	s.Equal(int64(3), total)
	s.Len(products, 3)

	// Verify order is correct (oldest first)
	s.Equal("Monthly Plan", products[0].Name)
	s.Equal("Yearly Plan", products[1].Name)
	s.Equal("Lifetime Plan", products[2].Name)
}

func (s *ProductRepositoryTestSuite) TestGetProduct() {
	tests := []struct {
		name          string
		id            string
		expectError   bool
		expectedError string
	}{
		{
			name:        "Valid existing product",
			id:          "11111111-1111-1111-1111-111111111111",
			expectError: false,
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
				s.Equal("Monthly Plan", product.Name)
				s.Equal(9.99, product.Price)
			}
		})
	}
}
