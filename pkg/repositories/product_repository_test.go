package repositories_test

import (
	"gymondo_dz/pkg/models"
	"gymondo_dz/pkg/repositories"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProductRepository(t *testing.T) {
	repo := repositories.NewProductRepository()

	validID := "11111111-1111-1111-1111-111111111111"
	invalidID := "00000000-0000-0000-0000-000000000000"
	malformedID := "not-a-uuid"

	tests := []struct {
		name          string
		method        func() (interface{}, error)
		expectError   bool
		expectedError string
		validate      func(t *testing.T, result interface{})
	}{
		// ------------------------- GetProducts
		{
			name: "GetProducts returns all products",
			method: func() (interface{}, error) {
				return repo.GetProducts()
			},
			expectError: false,
			validate: func(t *testing.T, result interface{}) {
				products := result.([]models.Product)
				assert.Len(t, products, 3)
				assert.Equal(t, "Monthly Plan", products[0].Name)
				assert.Equal(t, models.DurationMonth, products[0].Duration)
			},
		},
		// -------------------------- GetProduct
		{
			name: "GetProduct returns correct product for valid UUID",
			method: func() (interface{}, error) {
				return repo.GetProduct(validID)
			},
			expectError: false,
			validate: func(t *testing.T, result interface{}) {
				product := result.(*models.Product)
				assert.Equal(t, "Monthly Plan", product.Name)
				assert.Equal(t, 9.99, product.Price)
			},
		},
		{
			name: "GetProduct fails for non-existent UUID",
			method: func() (interface{}, error) {
				return repo.GetProduct(invalidID)
			},
			expectError:   true,
			expectedError: "product not found",
		},
		{
			name: "GetProduct fails for malformed UUID",
			method: func() (interface{}, error) {
				return repo.GetProduct(malformedID)
			},
			expectError:   true,
			expectedError: "invalid product ID format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.method()

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err.Error())
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, result)
				}
			}
		})
	}
}
