package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gymondo_dz/pkg/models"
	"gymondo_dz/pkg/repositories"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockProductRepository with custom error support
type MockProductRepository struct {
	mock.Mock
}

func (m *MockProductRepository) GetProducts() ([]models.Product, error) {
	args := m.Called()
	return args.Get(0).([]models.Product), args.Error(1)
}

func (m *MockProductRepository) GetProduct(id string) (*models.Product, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Product), args.Error(1)
}

func TestProductHandler(t *testing.T) {
	// Test products
	validProduct := &models.Product{
		ID:       uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		Name:     "Monthly Plan",
		Price:    9.99,
		Duration: models.DurationMonth,
	}

	tests := []struct {
		name           string
		endpoint       string
		mockSetup      func(*MockProductRepository)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:     "GetProducts - Success",
			endpoint: "/products",
			mockSetup: func(m *MockProductRepository) {
				m.On("GetProducts").Return([]models.Product{*validProduct}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"data":[{"id":"11111111-1111-1111-1111-111111111111","name":"Monthly Plan","description":"","price":9.99,"duration":30,"created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z"}]}`,
		},
		{
			name:     "GetProducts - Internal Error",
			endpoint: "/products",
			mockSetup: func(m *MockProductRepository) {
				m.On("GetProducts").Return([]models.Product{}, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"assert.AnError general error for testing"}`,
		},
		{
			name:     "GetProduct - Success",
			endpoint: "/products/11111111-1111-1111-1111-111111111111",
			mockSetup: func(m *MockProductRepository) {
				m.On("GetProduct", "11111111-1111-1111-1111-111111111111").Return(validProduct, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"data":{"id":"11111111-1111-1111-1111-111111111111","name":"Monthly Plan","description":"","price":9.99,"duration":30,"created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z"}}`,
		},
		{
			name:     "GetProduct - Not Found",
			endpoint: "/products/00000000-0000-0000-0000-000000000000",
			mockSetup: func(m *MockProductRepository) {
				m.On("GetProduct", "00000000-0000-0000-0000-000000000000").Return(nil, repositories.ErrProductNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"product not found"}`,
		},
		{
			name:     "GetProduct - Invalid UUID",
			endpoint: "/products/invalid-uuid",
			mockSetup: func(m *MockProductRepository) {
				m.On("GetProduct", "invalid-uuid").Return(nil, repositories.ErrInvalidProductID)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"details":"invalid product ID format","error":"invalid request"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			gin.SetMode(gin.TestMode)
			mockRepo := new(MockProductRepository)
			tt.mockSetup(mockRepo)
			handler := NewProductHandler(mockRepo)

			// Create router
			router := gin.New()
			router.GET("/products", handler.GetProducts)
			router.GET("/products/:id", handler.GetProduct)

			// Create request
			req := httptest.NewRequest("GET", tt.endpoint, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetProductsWithLargeDataset(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockRepo := new(MockProductRepository)

	// Generate 1000 products (large enough to test performance, small enough for unit tests)
	largeProducts := make([]models.Product, 1000)
	for i := 0; i < 1000; i++ {
		largeProducts[i] = models.Product{
			ID:        uuid.New(),
			Name:      fmt.Sprintf("Product %d", i),
			Price:     9.99 + float64(i),
			Duration:  models.DurationMonth,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	}

	mockRepo.On("GetProducts").Return(largeProducts, nil)
	handler := NewProductHandler(mockRepo)

	router := gin.New()
	router.GET("/products", handler.GetProducts)

	req := httptest.NewRequest("GET", "/products", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string][]models.Product
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Len(t, response["data"], 1000)
	assert.Equal(t, "Product 0", response["data"][0].Name)
	assert.Equal(t, "Product 999", response["data"][999].Name)
}
