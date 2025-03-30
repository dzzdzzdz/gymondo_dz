package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gymondo_dz/pkg/handlers"
	"gymondo_dz/pkg/models"
	"gymondo_dz/pkg/repositories"
	"gymondo_dz/pkg/testutils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestProductHandler(t *testing.T) {
	fixedTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	mockProduct := models.Product{
		ID:          uuid.MustParse("465dc700-666c-4b7a-80e2-d9e2967f4442"),
		Name:        "Test Product",
		Description: "Test Description",
		Price:       9.99,
		TaxRate:     0.10,
		Duration:    models.DurationMonth,
		CreatedAt:   fixedTime,
		UpdatedAt:   fixedTime,
	}

	// Apply AfterFind hook manually since we're mocking
	mockProduct.TotalPrice = mockProduct.Price + (mockProduct.Price * mockProduct.TaxRate)

	tests := []struct {
		name           string
		method         string
		path           string
		query          string
		mockSetup      func(*testutils.MockProductRepository)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "GetProducts success",
			method: "GET",
			path:   "/products",
			query:  "page=1&limit=10",
			mockSetup: func(m *testutils.MockProductRepository) {
				m.On("GetProducts", 1, 10).Return([]models.Product{mockProduct}, int64(1), nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"data":[{"id":"465dc700-666c-4b7a-80e2-d9e2967f4442","name":"Test Product","description":"Test Description","price":9.99,"tax_rate":0.1,"total_price":10.989,"duration":30,"created_at":"2025-01-01T00:00:00Z","updated_at":"2025-01-01T00:00:00Z"}],"meta":{"total":1,"page":1,"limit":10}}`,
		},
		{
			name:   "GetProducts default pagination",
			method: "GET",
			path:   "/products",
			query:  "",
			mockSetup: func(m *testutils.MockProductRepository) {
				m.On("GetProducts", 1, 10).Return([]models.Product{mockProduct}, int64(1), nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"data":[{"id":"465dc700-666c-4b7a-80e2-d9e2967f4442","name":"Test Product","description":"Test Description","price":9.99,"tax_rate":0.1,"total_price":10.989,"duration":30,"created_at":"2025-01-01T00:00:00Z","updated_at":"2025-01-01T00:00:00Z"}],"meta":{"total":1,"page":1,"limit":10}}`,
		},
		{
			name:   "GetProduct success",
			method: "GET",
			path:   "/products/465dc700-666c-4b7a-80e2-d9e2967f4442",
			mockSetup: func(m *testutils.MockProductRepository) {
				m.On("GetProduct", "465dc700-666c-4b7a-80e2-d9e2967f4442").Return(&mockProduct, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"data":{"id":"465dc700-666c-4b7a-80e2-d9e2967f4442","name":"Test Product","description":"Test Description","price":9.99,"tax_rate":0.1,"total_price":10.989,"duration":30,"created_at":"2025-01-01T00:00:00Z","updated_at":"2025-01-01T00:00:00Z"}}`,
		},
		{
			name:   "GetProduct not found",
			method: "GET",
			path:   "/products/465dc700-666c-4b7a-80e2-d9e2967f4442",
			mockSetup: func(m *testutils.MockProductRepository) {
				m.On("GetProduct", "465dc700-666c-4b7a-80e2-d9e2967f4442").Return(nil, repositories.ErrProductNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":{"message":"product not found","code":"not_found"}}`,
		},
		{
			name:   "GetProduct invalid UUID",
			method: "GET",
			path:   "/products/invalid-uuid",
			mockSetup: func(m *testutils.MockProductRepository) {
				// Expect the repository to return ErrInvalidProductID
				m.On("GetProduct", "invalid-uuid").Return(nil, repositories.ErrInvalidProductID)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":{"message":"invalid product ID","code":"invalid_id"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			mockRepo := new(testutils.MockProductRepository)
			tt.mockSetup(mockRepo)

			// Create handler and router
			handler := handlers.NewProductHandler(mockRepo)
			router := gin.Default()
			router.GET("/products", handler.GetProducts)
			router.GET("/products/:id", handler.GetProduct)

			// Create request with query parameters if any
			fullPath := tt.path
			if tt.query != "" {
				fullPath = tt.path + "?" + tt.query
			}
			req := httptest.NewRequest(tt.method, fullPath, nil)
			w := httptest.NewRecorder()

			// Serve request
			router.ServeHTTP(w, req)

			// Verify
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
			mockRepo.AssertExpectations(t)
		})
	}
}
