package handlers

import (
	"errors"
	"net/http"

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
	switch {
	case errors.Is(err, repositories.ErrSubscriptionNotFound):
		c.JSON(http.StatusNotFound, gin.H{
			"error": "subscription not found",
		})
	case errors.Is(err, repositories.ErrInvalidSubscriptionID):
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid subscription ID",
		})
	case err != nil:
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
	default:
		c.JSON(http.StatusOK, gin.H{
			"data": sub,
		})
	}
}

func (h *SubscriptionHandler) CreateSubscription(c *gin.Context) {
	productID := c.Param("product_id")

	// validate product exists
	product, err := h.productRepo.GetProduct(productID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid product ID",
		})
		return
	}

	// generating randomly for now
	userID := uuid.New().String()

	sub, err := h.repo.CreateSubscription(userID, product)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create subscription",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": sub,
	})
}

func (h *SubscriptionHandler) PauseSubscription(c *gin.Context) {
	subID := c.Param("id")

	sub, err := h.repo.PauseSubscription(subID)
	switch {
	case errors.Is(err, repositories.ErrCannotPause):
		c.JSON(http.StatusConflict, gin.H{
			"error": "subscription cannot be paused",
		})
	case errors.Is(err, repositories.ErrSubscriptionNotFound):
		c.JSON(http.StatusNotFound, gin.H{
			"error": "subscription not found",
		})
	case err != nil:
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
	default:
		c.JSON(http.StatusOK, gin.H{
			"data": sub,
		})
	}
}

func (h *SubscriptionHandler) UnpauseSubscription(c *gin.Context) {
	subID := c.Param("id")

	sub, err := h.repo.UnpauseSubscription(subID)
	switch {
	case errors.Is(err, repositories.ErrCannotUnpause):
		c.JSON(http.StatusConflict, gin.H{
			"error": "subscription cannot be unpaused",
		})
	case errors.Is(err, repositories.ErrSubscriptionNotFound):
		c.JSON(http.StatusNotFound, gin.H{
			"error": "subscription not found",
		})
	case err != nil:
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
	default:
		c.JSON(http.StatusOK, gin.H{
			"data": sub,
		})
	}
}

func (h *SubscriptionHandler) CancelSubscription(c *gin.Context) {
	subID := c.Param("id")

	sub, err := h.repo.CancelSubscription(subID)
	switch {
	case errors.Is(err, repositories.ErrCannotCancel):
		c.JSON(http.StatusConflict, gin.H{
			"error": "subscription cannot be cancelled",
		})
	case errors.Is(err, repositories.ErrSubscriptionNotFound):
		c.JSON(http.StatusNotFound, gin.H{
			"error": "subscription not found",
		})
	case err != nil:
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
	default:
		c.JSON(http.StatusOK, gin.H{
			"data": sub,
		})
	}
}
