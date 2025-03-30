package repositories

import (
	"errors"
	"gymondo_dz/pkg/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrProductNotFound  = errors.New("product not found")
	ErrInvalidProductID = errors.New("invalid product ID format")
)

type ProductRepository interface {
	GetProducts() ([]models.Product, error)
	GetProduct(id string) (*models.Product, error)
}

type ProductRepositoryImpl struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) ProductRepository {
	return &ProductRepositoryImpl{db: db}
}

func (r *ProductRepositoryImpl) GetProducts() ([]models.Product, error) {
	var products []models.Product
	result := r.db.Find(&products)
	if result.Error != nil {
		return nil, result.Error
	}
	return products, nil
}

func (r *ProductRepositoryImpl) GetProduct(id string) (*models.Product, error) {
	productID, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrInvalidProductID
	}

	var product models.Product
	result := r.db.First(&product, "id = ?", productID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, result.Error
	}

	return &product, nil
}
