package repositories_test

import (
	"gymondo_dz/pkg/models"
	"gymondo_dz/pkg/repositories"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
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
	s.db, err = gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		SkipDefaultTransaction: true,
	})
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
	// Clear existing data completely (including soft deleted)
	if err := s.db.Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Product{}).Error; err != nil {
		s.FailNow("Failed to clear products: " + err.Error())
	}

	// Create test products with all required fields
	now := time.Now().UTC()
	testProducts := []*models.Product{
		{
			ID:          uuid.MustParse("11111111-1111-1111-1111-111111111111"),
			Name:        "Monthly Plan",
			Description: "1 month subscription",
			Price:       9.99,
			TaxRate:     0.10, // Ensure this matches your model's default
			Duration:    models.DurationMonth,
			CreatedAt:   now.Add(-3 * time.Hour),
			UpdatedAt:   now.Add(-3 * time.Hour),
			DeletedAt:   gorm.DeletedAt{}, // Explicitly set to not deleted
		},
		{
			ID:          uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			Name:        "Yearly Plan",
			Description: "1 year subscription",
			Price:       99.99,
			TaxRate:     0.10,
			Duration:    models.DurationYear,
			CreatedAt:   now.Add(-2 * time.Hour),
			UpdatedAt:   now.Add(-2 * time.Hour),
			DeletedAt:   gorm.DeletedAt{},
		},
		{
			ID:          uuid.MustParse("33333333-3333-3333-3333-333333333333"),
			Name:        "Lifetime Plan",
			Description: "Lifetime access",
			Price:       999.99,
			TaxRate:     0.10,
			Duration:    models.DurationLifetime,
			CreatedAt:   now.Add(-1 * time.Hour),
			UpdatedAt:   now.Add(-1 * time.Hour),
			DeletedAt:   gorm.DeletedAt{},
		},
	}

	// Create products using direct SQL to bypass any hooks
	for _, p := range testProducts {
		result := s.db.Exec(`
			INSERT INTO products (id, name, description, price, tax_rate, duration, created_at, updated_at, deleted_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			p.ID, p.Name, p.Description, p.Price, p.TaxRate, p.Duration, p.CreatedAt, p.UpdatedAt, nil,
		)
		if result.Error != nil {
			s.FailNow("Failed to seed test data: " + result.Error.Error())
		}
	}

	// Verify data was inserted
	var count int64
	s.db.Model(&models.Product{}).Count(&count)
	assert.Equal(s.T(), int64(3), count, "Should have 3 products in database")
}

func TestProductRepositorySuite(t *testing.T) {
	suite.Run(t, new(ProductRepositoryTestSuite))
}

func (s *ProductRepositoryTestSuite) TestGetProducts() {
	products, total, err := s.repo.GetProducts(1, 10)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), int64(3), total)
	assert.Len(s.T(), products, 3)

	// Verify order is correct (oldest first)
	assert.Equal(s.T(), "Monthly Plan", products[0].Name)
	assert.Equal(s.T(), "Yearly Plan", products[1].Name)
	assert.Equal(s.T(), "Lifetime Plan", products[2].Name)
}

func (s *ProductRepositoryTestSuite) TestGetProductsPerformance() {
	// Seed large dataset
	for i := 0; i < 1000; i++ {
		product := models.Product{
			Name:  "Product " + strconv.Itoa(i),
			Price: float64(i),
		}
		s.NoError(s.db.Create(&product).Error)
	}

	start := time.Now()
	_, _, err := s.repo.GetProducts(1, 100)
	s.NoError(err)
	s.True(time.Since(start) < time.Second, "Pagination query too slow")
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
				assert.Error(s.T(), err)
				assert.Equal(s.T(), tt.expectedError, err.Error())
				assert.Nil(s.T(), product)
			} else {
				assert.NoError(s.T(), err)
				assert.NotNil(s.T(), product)
				assert.Equal(s.T(), "Monthly Plan", product.Name)
				assert.Equal(s.T(), 9.99, product.Price)
			}
		})
	}
}
