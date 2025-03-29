package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gymondo_dz/pkg/handlers"
	"gymondo_dz/pkg/models"
	"gymondo_dz/pkg/repositories"
	testutil "gymondo_dz/pkg/testutils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func setupProductRouter(h *handlers.ProductHandler) *gin.Engine {
	router := gin.Default()
	router.GET("/products", h.GetProducts)
	router.GET("/products/:id", h.GetProduct)
	return router
}

func TestProductHandler(t *testing.T) {
	fixedTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	mockProduct := &models.Product{
		ID:          uuid.MustParse("465dc700-666c-4b7a-80e2-d9e2967f4442"),
		Name:        "Test Product",
		Description: "",
		Price:       9.99,
		Duration:    30,
		CreatedAt:   fixedTime,
		UpdatedAt:   fixedTime,
	}
	mockProducts := []models.Product{*mockProduct}

	tests := []struct {
		name           string
		method         string
		path           string
		mockSetup      func(*testutil.MockProductRepository)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "GetProducts success",
			method: "GET",
			path:   "/products",
			mockSetup: func(m *testutil.MockProductRepository) {
				m.On("GetProducts").Return(mockProducts, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"data":[{"id":"465dc700-666c-4b7a-80e2-d9e2967f4442","name":"Test Product","description":"","price":9.99,"duration":30,"created_at":"2025-01-01T00:00:00Z","updated_at":"2025-01-01T00:00:00Z"}]}`,
		},
		{
			name:   "GetProducts empty",
			method: "GET",
			path:   "/products",
			mockSetup: func(m *testutil.MockProductRepository) {
				m.On("GetProducts").Return([]models.Product{}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"data":[]}`,
		},
		{
			name:   "GetProducts error",
			method: "GET",
			path:   "/products",
			mockSetup: func(m *testutil.MockProductRepository) {
				m.On("GetProducts").Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"assert.AnError general error for testing"}`,
		},
		{
			name:   "GetProduct success",
			method: "GET",
			path:   "/products/465dc700-666c-4b7a-80e2-d9e2967f4442",
			mockSetup: func(m *testutil.MockProductRepository) {
				m.On("GetProduct", "465dc700-666c-4b7a-80e2-d9e2967f4442").Return(mockProduct, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"data":{"id":"465dc700-666c-4b7a-80e2-d9e2967f4442","name":"Test Product","description":"","price":9.99,"duration":30,"created_at":"2025-01-01T00:00:00Z","updated_at":"2025-01-01T00:00:00Z"}}`,
		},
		{
			name:   "GetProduct not found",
			method: "GET",
			path:   "/products/465dc700-666c-4b7a-80e2-d9e2967f4442",
			mockSetup: func(m *testutil.MockProductRepository) {
				m.On("GetProduct", "465dc700-666c-4b7a-80e2-d9e2967f4442").Return(nil, repositories.ErrProductNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"product not found"}`,
		},
		{
			name:   "GetProduct invalid ID",
			method: "GET",
			path:   "/products/invalid",
			mockSetup: func(m *testutil.MockProductRepository) {
				m.On("GetProduct", "invalid").Return(nil, repositories.ErrInvalidProductID)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"details":"invalid product ID format", "error":"invalid request"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(testutil.MockProductRepository)
			tt.mockSetup(mockRepo)

			handler := handlers.NewProductHandler(mockRepo)
			router := setupProductRouter(handler)

			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
			mockRepo.AssertExpectations(t)
		})
	}
}
