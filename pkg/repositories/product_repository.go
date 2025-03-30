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
	GetProducts(page, limit int) ([]models.Product, int64, error)
	GetProduct(id string) (*models.Product, error)
}

type Pagination struct {
	Page  int
	Limit int
	Total int64
}

type ProductRepositoryImpl struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) ProductRepository {
	return &ProductRepositoryImpl{db: db}
}

func (r *ProductRepositoryImpl) GetProducts(page, limit int) ([]models.Product, int64, error) {
	var products []models.Product
	var total int64

	if err := r.db.Model(&models.Product{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if page < 1 {
		page = 1
	}

	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	result := r.db.Order("created_at ASC").Offset(offset).Limit(limit).Find(&products)
	if result.Error != nil {
		return nil, 0, result.Error
	}
	return products, total, nil
}

func (r *ProductRepositoryImpl) GetProduct(id string) (*models.Product, error) {
	productID, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrInvalidProductID
	}

	var product models.Product
	result := r.db.Debug().Where("id = ?", productID).First(&product)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, result.Error
	}

	return &product, nil
}
