package handlers

import (
	"errors"
	"net/http"

	"gymondo_dz/pkg/api"
	"gymondo_dz/pkg/repositories"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SubscriptionHandler struct {
	repo        repositories.SubscriptionRepository
	productRepo repositories.ProductRepository
}

func NewSubscriptionHandler(
	repo repositories.SubscriptionRepository,
	productRepo repositories.ProductRepository,
) *SubscriptionHandler {
	return &SubscriptionHandler{
		repo:        repo,
		productRepo: productRepo,
	}
}

func (h *SubscriptionHandler) GetSubscription(c *gin.Context) {
	subID := c.Param("id")

	sub, err := h.repo.GetSubscription(subID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, api.SuccessResponse(sub, nil))
}

func (h *SubscriptionHandler) CreateSubscription(c *gin.Context) {
	productID := c.Param("product_id")

	product, err := h.productRepo.GetProduct(productID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// In a real app, this would come from auth context
	userID := uuid.New().String()

	sub, err := h.repo.CreateSubscription(userID, product)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, api.SuccessResponse(sub, nil))
}

func (h *SubscriptionHandler) PauseSubscription(c *gin.Context) {
	subID := c.Param("id")

	sub, err := h.repo.PauseSubscription(subID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, api.SuccessResponse(sub, nil))
}

func (h *SubscriptionHandler) UnpauseSubscription(c *gin.Context) {
	subID := c.Param("id")

	sub, err := h.repo.UnpauseSubscription(subID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, api.SuccessResponse(sub, nil))
}

func (h *SubscriptionHandler) CancelSubscription(c *gin.Context) {
	subID := c.Param("id")

	sub, err := h.repo.CancelSubscription(subID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, api.SuccessResponse(sub, nil))
}

func (h *SubscriptionHandler) handleError(c *gin.Context, err error) {
	var status int
	var message, code string

	switch {
	case errors.Is(err, repositories.ErrSubscriptionNotFound),
		errors.Is(err, repositories.ErrProductNotFound):
		status = http.StatusNotFound
		message = "resource not found"
		code = "not_found"
	case errors.Is(err, repositories.ErrInvalidSubscriptionID),
		errors.Is(err, repositories.ErrInvalidProductID):
		status = http.StatusBadRequest
		message = "invalid ID format"
		code = "invalid_id"
	case errors.Is(err, repositories.ErrCannotPause):
		status = http.StatusConflict
		message = "cannot pause subscription"
		code = "invalid_state"
	case errors.Is(err, repositories.ErrCannotUnpause):
		status = http.StatusConflict
		message = "cannot unpause subscription"
		code = "invalid_state"
	case errors.Is(err, repositories.ErrCannotCancel):
		status = http.StatusConflict
		message = "cannot cancel subscription"
		code = "invalid_state"
	default:
		status = http.StatusInternalServerError
		message = "internal server error"
		code = "internal_error"
	}

	c.JSON(status, api.ErrorResponse(message, code))
	c.Abort() // Prevent any further handlers from being called
}
