package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"gymondo_dz/pkg/api"
	"gymondo_dz/pkg/handlers"
	"gymondo_dz/pkg/models"
	"gymondo_dz/pkg/repositories"
	"gymondo_dz/pkg/testutils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	now := time.Now()
	validProduct := &models.Product{
		ID:       uuid.New(),
		Name:     "Test Product",
		Duration: 30,
		Price:    9.99,
	}

	activeSub := &models.Subscription{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		ProductID: validProduct.ID,
		Status:    models.StatusActive,
		StartDate: now,
		EndDate:   now.Add(30 * 24 * time.Hour),
	}

	pausedSub := &models.Subscription{
		ID:        activeSub.ID,
		UserID:    activeSub.UserID,
		ProductID: activeSub.ProductID,
		Status:    models.StatusPaused,
		StartDate: activeSub.StartDate,
		EndDate:   activeSub.EndDate,
		PausedAt:  &now,
	}

	cancelledSub := &models.Subscription{
		ID:          activeSub.ID,
		UserID:      activeSub.UserID,
		ProductID:   activeSub.ProductID,
		Status:      models.StatusCancelled,
		StartDate:   activeSub.StartDate,
		EndDate:     activeSub.EndDate,
		CancelledAt: &now,
	}

	t.Run("Create Subscription - Success", func(t *testing.T) {
		mockProductRepo := new(testutils.MockProductRepository)
		mockSubRepo := new(testutils.MockSubscriptionRepository)

		mockProductRepo.On("GetProduct", validProduct.ID.String()).Return(validProduct, nil)
		mockSubRepo.On("CreateSubscription", mock.Anything, validProduct).Return(activeSub, nil)

		handler := handlers.NewSubscriptionHandler(mockSubRepo, mockProductRepo)
		router := setupSubscriptionRouter(handler)

		req := httptest.NewRequest("POST", "/products/"+validProduct.ID.String()+"/subscriptions", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response api.Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		responseData := response.Data.(map[string]interface{})
		responseID, err := uuid.Parse(responseData["id"].(string))
		assert.NoError(t, err)

		assert.Equal(t, activeSub.ID, responseID)

		mockProductRepo.AssertExpectations(t)
		mockSubRepo.AssertExpectations(t)
	})

	t.Run("Get Subscription - Success", func(t *testing.T) {
		mockProductRepo := new(testutils.MockProductRepository)
		mockSubRepo := new(testutils.MockSubscriptionRepository)

		mockSubRepo.On("GetSubscription", activeSub.ID.String()).Return(activeSub, nil)

		handler := handlers.NewSubscriptionHandler(mockSubRepo, mockProductRepo)
		router := setupSubscriptionRouter(handler)

		req := httptest.NewRequest("GET", "/subscriptions/"+activeSub.ID.String(), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response api.Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		responseData := response.Data.(map[string]interface{})
		responseID, err := uuid.Parse(responseData["id"].(string))
		assert.NoError(t, err)

		assert.Equal(t, activeSub.ID, responseID)

		mockSubRepo.AssertExpectations(t)
	})

	t.Run("Pause Subscription - Success", func(t *testing.T) {
		mockProductRepo := new(testutils.MockProductRepository)
		mockSubRepo := new(testutils.MockSubscriptionRepository)

		expectedVersion := 1
		mockSubRepo.On("PauseSubscription", activeSub.ID.String(), expectedVersion).Return(pausedSub, nil)

		handler := handlers.NewSubscriptionHandler(mockSubRepo, mockProductRepo)
		router := setupSubscriptionRouter(handler)

		req := httptest.NewRequest("PATCH", "/subscriptions/"+activeSub.ID.String()+"/pause", nil)
		req.Header.Set("If-Match", strconv.Itoa(expectedVersion))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response api.Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		responseData := response.Data.(map[string]interface{})
		assert.Equal(t, string(models.StatusPaused), responseData["status"].(string))

		mockSubRepo.AssertExpectations(t)
	})

	t.Run("Create with Invalid Product ID", func(t *testing.T) {
		mockProductRepo := new(testutils.MockProductRepository)
		mockSubRepo := new(testutils.MockSubscriptionRepository)

		invalidID := "invalid-uuid"
		mockProductRepo.On("GetProduct", invalidID).Return(nil, repositories.ErrInvalidProductID)

		handler := handlers.NewSubscriptionHandler(mockSubRepo, mockProductRepo)
		router := setupSubscriptionRouter(handler)

		req := httptest.NewRequest("POST", "/products/"+invalidID+"/subscriptions", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response api.Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "invalid ID format", response.Error.Message)
		assert.Equal(t, "invalid_id", response.Error.Code)

		mockProductRepo.AssertExpectations(t)
	})

	t.Run("Pause Already Cancelled Subscription", func(t *testing.T) {
		mockProductRepo := new(testutils.MockProductRepository)
		mockSubRepo := new(testutils.MockSubscriptionRepository)

		expectedVersion := 1
		mockSubRepo.On("PauseSubscription", cancelledSub.ID.String(), expectedVersion).Return(nil, repositories.ErrCannotPause)

		handler := handlers.NewSubscriptionHandler(mockSubRepo, mockProductRepo)
		router := setupSubscriptionRouter(handler)

		req := httptest.NewRequest("PATCH", "/subscriptions/"+cancelledSub.ID.String()+"/pause", nil)
		req.Header.Set("If-Match", strconv.Itoa(expectedVersion))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)

		var response api.Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "cannot pause subscription", response.Error.Message)
		assert.Equal(t, "invalid_state", response.Error.Code)

		mockSubRepo.AssertExpectations(t)
	})

	t.Run("Pause Subscription - Missing Version", func(t *testing.T) {
		mockProductRepo := new(testutils.MockProductRepository)
		mockSubRepo := new(testutils.MockSubscriptionRepository)

		handler := handlers.NewSubscriptionHandler(mockSubRepo, mockProductRepo)
		router := setupSubscriptionRouter(handler)

		req := httptest.NewRequest("PATCH", "/subscriptions/"+activeSub.ID.String()+"/pause", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusPreconditionRequired, w.Code)
	})
}
