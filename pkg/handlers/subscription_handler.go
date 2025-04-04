package handlers

import (
	"errors"
	"net/http"
	"strconv"

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

// @Summary Create a new subscription
// @Description Create subscription for a product
// @Tags subscriptions
// @Accept  json
// @Produce  json
// @Param product_id path string true "Product ID"
// @Success 201 {object} api.Response{data=models.Subscription}
// @Failure 400 {object} api.Response
// @Failure 404 {object} api.Response
// @Failure 500 {object} api.Response
// @Router /subscriptions/{product_id} [post]
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

// @Summary Get subscription details
// @Description Get subscription by ID
// @Tags subscriptions
// @Produce  json
// @Param id path string true "Subscription ID"
// @Success 200 {object} api.Response{data=models.Subscription}
// @Failure 400 {object} api.Response
// @Failure 404 {object} api.Response
// @Router /subscriptions/{id} [get]
func (h *SubscriptionHandler) GetSubscription(c *gin.Context) {
	subID := c.Param("id")

	sub, err := h.repo.GetSubscription(subID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, api.SuccessResponse(sub, nil))
}

// @Summary Pause subscription
// @Description Pause subscription by ID
// @Tags subscriptions
// @Produce  json
// @Param id path string true "Subscription ID"
// @Success 200 {object} api.Response{data=models.Subscription}
// @Failure 400 {object} api.Response
// @Failure 404 {object} api.Response
// @Failure 409 {object} api.Response
// @Router /subscriptions/{id}/pause [patch]
func (h *SubscriptionHandler) PauseSubscription(c *gin.Context) {
	subID := c.Param("id")

	versionHeader := c.GetHeader("If-Match")
	if versionHeader == "" {
		c.JSON(http.StatusPreconditionRequired, gin.H{
			"error": "Missing If-Match header",
		})
		return
	}

	version, err := strconv.Atoi(versionHeader)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid If-Match header format",
		})
		return
	}

	sub, err := h.repo.PauseSubscription(subID, version)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, api.SuccessResponse(sub, nil))
}

// @Summary Unpause subscription
// @Description Unpause subscription by ID
// @Tags subscriptions
// @Produce  json
// @Param id path string true "Subscription ID"
// @Success 200 {object} api.Response{data=models.Subscription}
// @Failure 400 {object} api.Response
// @Failure 404 {object} api.Response
// @Failure 409 {object} api.Response
// @Router /subscriptions/{id}/unpause [patch]
func (h *SubscriptionHandler) UnpauseSubscription(c *gin.Context) {
	subID := c.Param("id")

	versionHeader := c.GetHeader("If-Match")
	if versionHeader == "" {
		c.JSON(http.StatusPreconditionRequired, gin.H{
			"error": "Missing If-Match header",
		})
		return
	}

	version, err := strconv.Atoi(versionHeader)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid If-Match header format",
		})
		return
	}

	sub, err := h.repo.UnpauseSubscription(subID, version)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, api.SuccessResponse(sub, nil))
}

// @Summary Cancel subscription
// @Description Cancel subscription by ID
// @Tags subscriptions
// @Produce  json
// @Param id path string true "Subscription ID"
// @Success 200 {object} api.Response{data=models.Subscription}
// @Failure 400 {object} api.Response
// @Failure 404 {object} api.Response
// @Failure 409 {object} api.Response
// @Router /subscriptions/{id} [delete]
func (h *SubscriptionHandler) CancelSubscription(c *gin.Context) {
	subID := c.Param("id")
	versionHeader := c.GetHeader("If-Match")
	if versionHeader == "" {
		c.JSON(http.StatusPreconditionRequired, gin.H{
			"error": "Missing If-Match header",
		})
		return
	}

	version, err := strconv.Atoi(versionHeader)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid If-Match header format",
		})
		return
	}

	sub, err := h.repo.CancelSubscription(subID, version)
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
	case errors.Is(err, repositories.ErrConcurrentModification):
		status = http.StatusConflict
		message = "subscription was modified by another request"
		code = "concurrent_modification"
	default:
		status = http.StatusInternalServerError
		message = "internal server error"
		code = "internal_error"
	}

	c.JSON(status, api.ErrorResponse(message, code))
	c.Abort() // Prevent any further handlers from being called
}
