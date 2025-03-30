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

func (m *MockProductRepository) GetProducts() ([]models.Product, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Product), args.Error(1)
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

func (m *MockSubscriptionRepository) PauseSubscription(id string) (*models.Subscription, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) UnpauseSubscription(id string) (*models.Subscription, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) CancelSubscription(id string) (*models.Subscription, error) {
	args := m.Called(id)
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
