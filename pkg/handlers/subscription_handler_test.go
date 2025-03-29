package handlers_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"gymondo_dz/pkg/handlers"
	"gymondo_dz/pkg/models"
	"gymondo_dz/pkg/repositories"
	testutil "gymondo_dz/pkg/testutils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupSubscriptionRouter(h *handlers.SubscriptionHandler) *gin.Engine {
	router := gin.Default()
	router.POST("/products/:product_id/subscriptions", h.CreateSubscription)
	router.GET("/subscriptions/:id", h.GetSubscription)
	router.PATCH("/subscriptions/:id/pause", h.PauseSubscription)
	router.PATCH("/subscriptions/:id/unpause", h.UnpauseSubscription)
	router.DELETE("/subscriptions/:id", h.CancelSubscription)
	return router
}

func TestSubscriptionHandler(t *testing.T) {
	validProduct := testutil.NewMockProduct()
	validSub := testutil.NewMockSubscription()
	cancelledSub := testutil.NewMockSubscription()
	cancelledSub.Status = models.StatusCancelled
	pausedSub := testutil.NewMockSubscription()
	pausedSub.Status = models.StatusPaused

	tests := []struct {
		name           string
		method         string
		path           string
		mockSetup      func(*testutil.MockProductRepository, *testutil.MockSubscriptionRepository)
		expectedStatus int
		expectedError  string
	}{
		{
			name:   "Create subscription success",
			method: "POST",
			path:   "/products/" + validProduct.ID.String() + "/subscriptions",
			mockSetup: func(p *testutil.MockProductRepository, s *testutil.MockSubscriptionRepository) {
				p.On("GetProduct", validProduct.ID.String()).Return(validProduct, nil)
				s.On("CreateSubscription", mock.Anything, validProduct).Return(validSub, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:   "Create with invalid product ID",
			method: "POST",
			path:   "/products/invalid/subscriptions",
			mockSetup: func(p *testutil.MockProductRepository, s *testutil.MockSubscriptionRepository) {
				p.On("GetProduct", "invalid").Return(nil, repositories.ErrInvalidProductID)
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid product ID",
		},
		{
			name:   "Create with repository error",
			method: "POST",
			path:   "/products/" + validProduct.ID.String() + "/subscriptions",
			mockSetup: func(p *testutil.MockProductRepository, s *testutil.MockSubscriptionRepository) {
				p.On("GetProduct", validProduct.ID.String()).Return(validProduct, nil)
				s.On("CreateSubscription", mock.Anything, validProduct).Return(nil, errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "failed to create subscription",
		},

		// GetSubscription tests
		{
			name:   "Get subscription success",
			method: "GET",
			path:   "/subscriptions/" + validSub.ID.String(),
			mockSetup: func(p *testutil.MockProductRepository, s *testutil.MockSubscriptionRepository) {
				s.On("GetSubscription", validSub.ID.String()).Return(validSub, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Get non-existent subscription",
			method: "GET",
			path:   "/subscriptions/" + validSub.ID.String(),
			mockSetup: func(p *testutil.MockProductRepository, s *testutil.MockSubscriptionRepository) {
				s.On("GetSubscription", validSub.ID.String()).Return(nil, repositories.ErrSubscriptionNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "subscription not found",
		},
		{
			name:   "Get with invalid ID format",
			method: "GET",
			path:   "/subscriptions/invalid",
			mockSetup: func(p *testutil.MockProductRepository, s *testutil.MockSubscriptionRepository) {
				s.On("GetSubscription", "invalid").Return(nil, repositories.ErrInvalidSubscriptionID)
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid subscription ID",
		},

		// PauseSubscription tests
		{
			name:   "Pause active subscription",
			method: "PATCH",
			path:   "/subscriptions/" + validSub.ID.String() + "/pause",
			mockSetup: func(p *testutil.MockProductRepository, s *testutil.MockSubscriptionRepository) {
				s.On("PauseSubscription", validSub.ID.String()).Return(pausedSub, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Pause cancelled subscription",
			method: "PATCH",
			path:   "/subscriptions/" + cancelledSub.ID.String() + "/pause",
			mockSetup: func(p *testutil.MockProductRepository, s *testutil.MockSubscriptionRepository) {
				s.On("PauseSubscription", cancelledSub.ID.String()).Return(nil, repositories.ErrCannotPause)
			},
			expectedStatus: http.StatusConflict,
			expectedError:  "subscription cannot be paused",
		},
		{
			name:   "Pause expired subscription",
			method: "PATCH",
			path:   "/subscriptions/expired/pause",
			mockSetup: func(p *testutil.MockProductRepository, s *testutil.MockSubscriptionRepository) {
				expiredSub := testutil.NewMockSubscription()
				expiredSub.Status = models.StatusExpired
				s.On("PauseSubscription", "expired").Return(nil, repositories.ErrCannotPause)
			},
			expectedStatus: http.StatusConflict,
			expectedError:  "subscription cannot be paused",
		},

		// UnpauseSubscription tests
		{
			name:   "Unpause paused subscription",
			method: "PATCH",
			path:   "/subscriptions/" + pausedSub.ID.String() + "/unpause",
			mockSetup: func(p *testutil.MockProductRepository, s *testutil.MockSubscriptionRepository) {
				s.On("UnpauseSubscription", pausedSub.ID.String()).Return(validSub, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Unpause active subscription",
			method: "PATCH",
			path:   "/subscriptions/" + validSub.ID.String() + "/unpause",
			mockSetup: func(p *testutil.MockProductRepository, s *testutil.MockSubscriptionRepository) {
				s.On("UnpauseSubscription", validSub.ID.String()).Return(nil, repositories.ErrCannotUnpause)
			},
			expectedStatus: http.StatusConflict,
			expectedError:  "subscription cannot be unpaused",
		},
		{
			name:   "Unpause expired subscription",
			method: "PATCH",
			path:   "/subscriptions/expired/unpause",
			mockSetup: func(p *testutil.MockProductRepository, s *testutil.MockSubscriptionRepository) {
				expiredSub := testutil.NewMockSubscription()
				expiredSub.Status = models.StatusExpired
				s.On("UnpauseSubscription", "expired").Return(nil, repositories.ErrCannotUnpause)
			},
			expectedStatus: http.StatusConflict,
			expectedError:  "subscription cannot be unpaused",
		},

		// CancelSubscription tests
		{
			name:   "Cancel active subscription",
			method: "DELETE",
			path:   "/subscriptions/" + validSub.ID.String(),
			mockSetup: func(p *testutil.MockProductRepository, s *testutil.MockSubscriptionRepository) {
				s.On("CancelSubscription", validSub.ID.String()).Return(cancelledSub, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Cancel already cancelled subscription",
			method: "DELETE",
			path:   "/subscriptions/" + cancelledSub.ID.String(),
			mockSetup: func(p *testutil.MockProductRepository, s *testutil.MockSubscriptionRepository) {
				s.On("CancelSubscription", cancelledSub.ID.String()).Return(nil, repositories.ErrCannotCancel)
			},
			expectedStatus: http.StatusConflict,
			expectedError:  "subscription cannot be cancelled",
		},
		{
			name:   "Cancel expired subscription",
			method: "DELETE",
			path:   "/subscriptions/expired",
			mockSetup: func(p *testutil.MockProductRepository, s *testutil.MockSubscriptionRepository) {
				expiredSub := testutil.NewMockSubscription()
				expiredSub.Status = models.StatusExpired
				s.On("CancelSubscription", "expired").Return(nil, repositories.ErrCannotCancel)
			},
			expectedStatus: http.StatusConflict,
			expectedError:  "subscription cannot be cancelled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProductRepo := new(testutil.MockProductRepository)
			mockSubRepo := new(testutil.MockSubscriptionRepository)

			tt.mockSetup(mockProductRepo, mockSubRepo)

			handler := handlers.NewSubscriptionHandler(mockSubRepo, mockProductRepo)
			router := setupSubscriptionRouter(handler)

			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, response["error"])
			}

			mockProductRepo.AssertExpectations(t)
			mockSubRepo.AssertExpectations(t)
		})
	}
}
