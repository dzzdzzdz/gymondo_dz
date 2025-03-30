package repositories_test

import (
	"testing"
	"time"

	"gymondo_dz/pkg/database"
	"gymondo_dz/pkg/models"
	"gymondo_dz/pkg/repositories"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type SubscriptionRepositoryTestSuite struct {
	suite.Suite
	db          *gorm.DB
	productRepo repositories.ProductRepository
	subRepo     repositories.SubscriptionRepository
}

func (s *SubscriptionRepositoryTestSuite) SetupSuite() {
	// Setup in-memory SQLite database
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		s.FailNow("Failed to connect to test database")
	}

	// Run migrations
	if err := database.AutoMigrate(db, true); err != nil {
		s.FailNow("Failed to migrate test database")
	}

	s.db = db
	s.productRepo = repositories.NewProductRepository(db)
	s.subRepo = repositories.NewSubscriptionRepository(db)
}

func (s *SubscriptionRepositoryTestSuite) SetupTest() {
	// Clear all data before each test
	s.db.Exec("DELETE FROM subscriptions")
	s.db.Exec("DELETE FROM products")
}

func TestSubscriptionRepositorySuite(t *testing.T) {
	suite.Run(t, new(SubscriptionRepositoryTestSuite))
}

func (s *SubscriptionRepositoryTestSuite) seedTestProduct() *models.Product {
	product := &models.Product{
		ID:       uuid.New(),
		Name:     "Test Product",
		Duration: models.DurationMonth,
		Price:    9.99,
	}
	s.NoError(s.db.Create(product).Error)
	return product
}

func (s *SubscriptionRepositoryTestSuite) TestCreateSubscription() {
	product := s.seedTestProduct()
	userID := uuid.New().String()

	// Test valid creation
	sub, err := s.subRepo.CreateSubscription(userID, product)
	s.NoError(err)
	s.NotNil(sub)
	s.Equal(userID, sub.UserID.String())
	s.Equal(product.ID, sub.ProductID)
	s.Equal(models.StatusActive, sub.Status)
	s.WithinDuration(time.Now(), sub.StartDate, time.Second)
	s.WithinDuration(time.Now().Add(time.Hour*24*time.Duration(product.Duration)), sub.EndDate, time.Second)

	// Test error cases
	tests := []struct {
		name          string
		userID        string
		product       *models.Product
		expectedError error
	}{
		{
			name:          "Invalid user ID",
			userID:        "invalid-uuid",
			product:       product,
			expectedError: repositories.ErrInvalidSubscriptionID,
		},
		{
			name:          "Nil product",
			userID:        userID,
			product:       nil,
			expectedError: repositories.ErrProductRequired,
		},
		{
			name:          "Invalid product duration",
			userID:        userID,
			product:       &models.Product{Duration: 0},
			expectedError: repositories.ErrInvalidProductDuration,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			sub, err := s.subRepo.CreateSubscription(tt.userID, tt.product)
			s.Error(err)
			s.Equal(tt.expectedError, err)
			s.Nil(sub)
		})
	}
}

func (s *SubscriptionRepositoryTestSuite) TestGetSubscription() {
	product := s.seedTestProduct()
	userID := uuid.New().String()

	// Create test subscription
	sub, err := s.subRepo.CreateSubscription(userID, product)
	s.NoError(err)

	// Test successful get
	retrieved, err := s.subRepo.GetSubscription(sub.ID.String())
	s.NoError(err)
	s.Equal(sub.ID, retrieved.ID)

	// Test not found
	_, err = s.subRepo.GetSubscription(uuid.New().String())
	s.Error(err)
	s.Equal(repositories.ErrSubscriptionNotFound, err)

	// Test invalid ID
	_, err = s.subRepo.GetSubscription("invalid-uuid")
	s.Error(err)
	s.Equal(repositories.ErrInvalidSubscriptionID, err)
}

func (s *SubscriptionRepositoryTestSuite) TestPauseUnpauseSubscription() {
	product := s.seedTestProduct()
	userID := uuid.New().String()
	sub, err := s.subRepo.CreateSubscription(userID, product)
	s.NoError(err)
	s.Equal(1, sub.Version)

	// Test pause with correct version
	pausedSub, err := s.subRepo.PauseSubscription(sub.ID.String(), sub.Version)
	s.NoError(err)
	s.Equal(models.StatusPaused, pausedSub.Status)
	s.NotNil(pausedSub.PausedAt)
	s.Equal(2, pausedSub.Version)

	// Test cannot pause with stale version
	_, err = s.subRepo.PauseSubscription(sub.ID.String(), 1)
	s.Error(err)
	s.Equal(repositories.ErrConcurrentModification, err)

	// Test cannot pause already paused (even with correct version)
	_, err = s.subRepo.PauseSubscription(sub.ID.String(), 2)
	s.Error(err)
	s.Equal(repositories.ErrCannotPause, err)

	// Test unpause with correct version
	unpausedSub, err := s.subRepo.UnpauseSubscription(sub.ID.String(), 2)
	s.NoError(err)
	s.Equal(models.StatusActive, unpausedSub.Status)
	s.Nil(unpausedSub.PausedAt)
	s.Equal(3, unpausedSub.Version)

	// Test cannot unpause with stale version
	_, err = s.subRepo.UnpauseSubscription(sub.ID.String(), 2)
	s.Error(err)
	s.Equal(repositories.ErrConcurrentModification, err)

	// Test cannot unpause active (even with correct version)
	_, err = s.subRepo.UnpauseSubscription(sub.ID.String(), 3)
	s.Error(err)
	s.Equal(repositories.ErrCannotUnpause, err)
}

func (s *SubscriptionRepositoryTestSuite) TestCancelSubscription() {
	product := s.seedTestProduct()
	userID := uuid.New().String()
	sub, err := s.subRepo.CreateSubscription(userID, product)
	s.NoError(err)
	s.Equal(1, sub.Version)

	// Test cancel with correct version
	cancelledSub, err := s.subRepo.CancelSubscription(sub.ID.String(), 1)
	s.NoError(err)
	s.Equal(models.StatusCancelled, cancelledSub.Status)
	s.NotNil(cancelledSub.CancelledAt)
	s.Equal(2, cancelledSub.Version)

	// Test cannot cancel with stale version
	_, err = s.subRepo.CancelSubscription(sub.ID.String(), 1)
	s.Error(err)
	s.Equal(repositories.ErrConcurrentModification, err)

	// Test cannot cancel already cancelled (even with correct version)
	_, err = s.subRepo.CancelSubscription(sub.ID.String(), 2)
	s.Error(err)
	s.Equal(repositories.ErrCannotCancel, err)
}

func (s *SubscriptionRepositoryTestSuite) TestAutoExpiration() {
	product := s.seedTestProduct()
	userID := uuid.New().String()
	sub, err := s.subRepo.CreateSubscription(userID, product)
	s.NoError(err)
	originalVersion := sub.Version

	// Manually set end date to past
	s.db.Model(&models.Subscription{}).Where("id = ?", sub.ID).
		Updates(map[string]interface{}{
			"end_date": time.Now().Add(-24 * time.Hour),
			"version":  originalVersion,
		})

	// Test auto-expiration on get
	retrieved, err := s.subRepo.GetSubscription(sub.ID.String())
	s.NoError(err)
	s.Equal(models.StatusExpired, retrieved.Status)
	s.Equal(originalVersion+1, retrieved.Version)
}

func (s *SubscriptionRepositoryTestSuite) TestUnpauseExtendsSubscription() {
	product := s.seedTestProduct()
	userID := uuid.New().String()
	sub, err := s.subRepo.CreateSubscription(userID, product)
	s.NoError(err)
	s.Equal(1, sub.Version)

	// Pause the subscription and record time
	beforePause := time.Now()
	pausedSub, err := s.subRepo.PauseSubscription(sub.ID.String(), 1)
	s.NoError(err)
	s.Equal(2, pausedSub.Version)

	// Verify pause happened after our marker time
	s.True(pausedSub.PausedAt.After(beforePause) || pausedSub.PausedAt.Equal(beforePause))

	// Calculate remaining duration when paused
	remainingDuration := pausedSub.EndDate.Sub(*pausedSub.PausedAt)

	// Wait a bit before unpausing to test time extension
	time.Sleep(100 * time.Millisecond)
	beforeUnpause := time.Now()

	// Unpause
	unpausedSub, err := s.subRepo.UnpauseSubscription(sub.ID.String(), 2)
	s.NoError(err)
	s.Equal(3, unpausedSub.Version)

	// Verify end date was extended correctly
	expectedEnd := beforeUnpause.Add(remainingDuration)
	s.WithinDuration(expectedEnd, unpausedSub.EndDate, time.Second)
	s.Equal(models.StatusActive, unpausedSub.Status)
	s.Nil(unpausedSub.PausedAt)

	_, err = s.subRepo.UnpauseSubscription(pausedSub.ID.String(), 2) // stale version
	s.ErrorIs(err, repositories.ErrConcurrentModification)
}

func (s *SubscriptionRepositoryTestSuite) TestConcurrentUpdates() {
	product := s.seedTestProduct()
	userID := uuid.New().String()
	sub, _ := s.subRepo.CreateSubscription(userID, product)

	// Simulate concurrent update by modifying the version directly in DB
	s.db.Model(&models.Subscription{}).Where("id = ?", sub.ID).
		Update("version", sub.Version+1)

	// All operations should fail with ErrConcurrentModification
	_, err := s.subRepo.PauseSubscription(sub.ID.String(), sub.Version)
	s.ErrorIs(err, repositories.ErrConcurrentModification)

	_, err = s.subRepo.UnpauseSubscription(sub.ID.String(), sub.Version)
	s.ErrorIs(err, repositories.ErrConcurrentModification)

	_, err = s.subRepo.CancelSubscription(sub.ID.String(), sub.Version)
	s.ErrorIs(err, repositories.ErrConcurrentModification)
}
