package handlers

import (
	"errors"
	"net/http"
	"strconv"

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

// @Summary List all products
// @Description Get a list of all available subscription products
// @Tags products
// @Produce json
// @Param page query int false "Page number" default(1) minimum(1)
// @Param limit query int false "Items per page" default(10) minimum(1) maximum(100)
// @Success 200 {object} api.Response{data=[]models.Product,meta=api.Meta} "Paginated list of products"
// @Failure 500 {object} api.Response "Internal server error"
// @Router /products [get]
func (h *ProductHandler) GetProducts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	products, total, err := h.repo.GetProducts(page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponse("failed to fetch products", "product_error"))
		return
	}

	response := api.SuccessResponse(products, &api.Meta{
		Page:  page,
		Limit: limit,
		Total: total,
	})
	c.JSON(http.StatusOK, response)
}

// @Summary Get product details
// @Description Get details for a specific product
// @Tags products
// @Produce json
// @Param id path string true "Product ID" format(uuid) example("d337a556-6fd6-47b9-b07f-4e60b9a78d2c")
// @Success 200 {object} api.Response{data=models.Product} "Product details"
// @Failure 400 {object} api.Response "Invalid ID format"
// @Failure 404 {object} api.Response "Product not found"
// @Failure 500 {object} api.Response "Internal server error"
// @Router /products/{id} [get]
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
