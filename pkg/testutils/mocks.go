package testutils

import (
	"time"

	"gymondo_dz/pkg/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockProductRepository implements ProductRepository for testing
type MockProductRepository struct {
	mock.Mock
}

func (m *MockProductRepository) GetProducts(page, limit int) ([]models.Product, int64, error) {
	args := m.Called(page, limit)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]models.Product), args.Get(1).(int64), args.Error(2)
}

func (m *MockProductRepository) GetProduct(id string) (*models.Product, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Product), args.Error(1)
}

// MockSubscriptionRepository implements SubscriptionRepository for testing
type MockSubscriptionRepository struct {
	mock.Mock
}

func (m *MockSubscriptionRepository) GetSubscription(id string) (*models.Subscription, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) CreateSubscription(userID string, product *models.Product) (*models.Subscription, error) {
	args := m.Called(userID, product)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) PauseSubscription(id string, expectedVersion int) (*models.Subscription, error) {
	args := m.Called(id, expectedVersion)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) UnpauseSubscription(id string, expectedVersion int) (*models.Subscription, error) {
	args := m.Called(id, expectedVersion)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) CancelSubscription(id string, expectedVersion int) (*models.Subscription, error) {
	args := m.Called(id, expectedVersion)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Subscription), args.Error(1)
}

// Helper functions for testing
func NewMockProduct() *models.Product {
	return &models.Product{
		ID:        uuid.New(),
		Name:      "Test Product",
		Duration:  30,
		Price:     9.99,
		TaxRate:   0.10,
		CreatedAt: time.Now(),
	}
}

func NewMockSubscription() *models.Subscription {
	return &models.Subscription{
		ID:        uuid.New(),
		Status:    models.StatusActive,
		StartDate: time.Now(),
		EndDate:   time.Now().Add(30 * 24 * time.Hour),
		CreatedAt: time.Now(),
	}
}

func NewMockProductList(count int) []models.Product {
	products := make([]models.Product, count)
	for i := range count {
		products[i] = *NewMockProduct()
		products[i].Name = products[i].Name + " " + string(rune('A'+i))
		products[i].CreatedAt = time.Now().Add(-time.Duration(count-i) * time.Hour) // Ensure different creation times
	}
	return products
}
