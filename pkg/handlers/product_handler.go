package handlers

import (
	"errors"
	"net/http"

	"gymondo_dz/pkg/repositories"

	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	repo repositories.ProductRepository
}

func NewProductHandler(repo repositories.ProductRepository) *ProductHandler {
	return &ProductHandler{repo: repo}
}

func (h *ProductHandler) GetProducts(c *gin.Context) {
	products, err := h.repo.GetProducts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}) // TODO: improve errors
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": products, // wrapper so we can add metadata in the future
	})
}

func (h *ProductHandler) GetProduct(c *gin.Context) {
	productID := c.Param("id")
	product, err := h.repo.GetProduct(productID)

	switch {
	case errors.Is(err, repositories.ErrProductNotFound):
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
	case errors.Is(err, repositories.ErrInvalidProductID):
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request",
			"details": err.Error(),
		})
	case err != nil:
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
	default:
		c.JSON(http.StatusOK, gin.H{
			"data": product,
		})
	}
}
