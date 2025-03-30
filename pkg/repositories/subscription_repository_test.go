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

	// Test pause
	pausedSub, err := s.subRepo.PauseSubscription(sub.ID.String())
	s.NoError(err)
	s.Equal(models.StatusPaused, pausedSub.Status)
	s.NotNil(pausedSub.PausedAt)

	// Test cannot pause already paused
	_, err = s.subRepo.PauseSubscription(sub.ID.String())
	s.Error(err)
	s.Equal(repositories.ErrCannotPause, err)

	// Test unpause
	unpausedSub, err := s.subRepo.UnpauseSubscription(sub.ID.String())
	s.NoError(err)
	s.Equal(models.StatusActive, unpausedSub.Status)
	s.Nil(unpausedSub.PausedAt)

	// Test cannot unpause active
	_, err = s.subRepo.UnpauseSubscription(sub.ID.String())
	s.Error(err)
	s.Equal(repositories.ErrCannotUnpause, err)
}

func (s *SubscriptionRepositoryTestSuite) TestCancelSubscription() {
	product := s.seedTestProduct()
	userID := uuid.New().String()
	sub, err := s.subRepo.CreateSubscription(userID, product)
	s.NoError(err)

	// Test cancel
	cancelledSub, err := s.subRepo.CancelSubscription(sub.ID.String())
	s.NoError(err)
	s.Equal(models.StatusCancelled, cancelledSub.Status)
	s.NotNil(cancelledSub.CancelledAt)

	// Test cannot cancel already cancelled
	_, err = s.subRepo.CancelSubscription(sub.ID.String())
	s.Error(err)
	s.Equal(repositories.ErrCannotCancel, err)
}

func (s *SubscriptionRepositoryTestSuite) TestAutoExpiration() {
	product := s.seedTestProduct()
	userID := uuid.New().String()
	sub, err := s.subRepo.CreateSubscription(userID, product)
	s.NoError(err)

	// Manually set end date to past
	s.db.Model(&models.Subscription{}).Where("id = ?", sub.ID).
		Update("end_date", time.Now().Add(-24*time.Hour))

	// Test auto-expiration on get
	retrieved, err := s.subRepo.GetSubscription(sub.ID.String())
	s.NoError(err)
	s.Equal(models.StatusExpired, retrieved.Status)
}

func (s *SubscriptionRepositoryTestSuite) TestUnpauseExtendsSubscription() {
	product := s.seedTestProduct()
	userID := uuid.New().String()
	sub, err := s.subRepo.CreateSubscription(userID, product)
	s.NoError(err)

	// Pause the subscription and record time
	beforePause := time.Now()
	pausedSub, err := s.subRepo.PauseSubscription(sub.ID.String())
	s.NoError(err)

	// Verify pause happened after our marker time
	s.True(pausedSub.PausedAt.After(beforePause) || pausedSub.PausedAt.Equal(beforePause))

	// Calculate remaining duration when paused
	remainingDuration := pausedSub.EndDate.Sub(*pausedSub.PausedAt)

	// Wait a bit before unpausing to test time extension
	time.Sleep(100 * time.Millisecond)
	beforeUnpause := time.Now()

	// Unpause
	unpausedSub, err := s.subRepo.UnpauseSubscription(sub.ID.String())
	s.NoError(err)

	// Verify end date was extended correctly
	expectedEnd := beforeUnpause.Add(remainingDuration)
	s.WithinDuration(expectedEnd, unpausedSub.EndDate, time.Second)
	s.Equal(models.StatusActive, unpausedSub.Status)
	s.Nil(unpausedSub.PausedAt)
}
