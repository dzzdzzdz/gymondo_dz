package repositories

import (
	"errors"
	"gymondo_dz/pkg/models"
	"time"

	"github.com/google/uuid"
)

type ProductRepository interface {
	GetProducts() ([]models.Product, error)
	GetProduct(id string) (*models.Product, error)
}

type ProductRepositoryImpl struct {
	products   []models.Product
	productMap map[uuid.UUID]models.Product
}

func NewProductRepository() ProductRepository {
	// initialize with hard-coded products
	products := []models.Product{
		{
			ID:          uuid.MustParse("11111111-1111-1111-1111-111111111111"),
			Name:        "Monthly Plan",
			Description: "Access all features for 1 month",
			Price:       9.99,
			Duration:    models.DurationMonth,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			Name:        "Yearly Plan",
			Description: "Access all features for 1 year (15% discount)",
			Price:       99.99,
			Duration:    models.DurationYear,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.MustParse("33333333-3333-3333-3333-333333333333"),
			Name:        "Lifetime Plan",
			Description: "Lifetime access (one-time payment)",
			Price:       299.99,
			Duration:    models.DurationLifetime,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	productMap := make(map[uuid.UUID]models.Product)
	for _, p := range products {
		productMap[p.ID] = p
	}

	return &ProductRepositoryImpl{
		products:   products,
		productMap: productMap,
	}
}

func (r *ProductRepositoryImpl) GetProducts() ([]models.Product, error) {
	return r.products, nil
}

func (r *ProductRepositoryImpl) GetProduct(id string) (*models.Product, error) {
	productID, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("invalid product ID format")
	}

	if product, exists := r.productMap[productID]; exists {
		return &product, nil
	}

	return nil, errors.New("product not found")
}
