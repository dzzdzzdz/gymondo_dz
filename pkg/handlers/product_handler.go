package handlers

import (
	"errors"
	"net/http"

	"gymondo_dz/pkg/api"
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
		c.JSON(http.StatusInternalServerError, api.ErrorResponse("failed to fetch products", "product_error"))
		return
	}

	response := api.SuccessResponse(products, &api.Meta{
		Total: len(products),
	})
	c.JSON(http.StatusOK, response)
}

func (h *ProductHandler) GetProduct(c *gin.Context) {
	productID := c.Param("id")
	product, err := h.repo.GetProduct(productID)

	switch {
	case errors.Is(err, repositories.ErrProductNotFound):
		c.JSON(http.StatusNotFound, api.ErrorResponse("product not found", "not_found"))
	case errors.Is(err, repositories.ErrInvalidProductID):
		c.JSON(http.StatusBadRequest, api.ErrorResponse("invalid product ID", "invalid_id"))
	case err != nil:
		c.JSON(http.StatusInternalServerError, api.ErrorResponse("internal server error", "internal_error"))
	default:
		c.JSON(http.StatusOK, api.SuccessResponse(product, nil))
	}
}
