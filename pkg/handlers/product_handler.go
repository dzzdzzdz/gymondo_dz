package handlers

import (
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
	if err != nil {
		if err.Error() == "product not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "product not found",
			})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid product ID",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": product,
	})
}
