package repositories_test

import (
	"testing"
	"time"

	"gymondo_dz/pkg/models"
	"gymondo_dz/pkg/repositories"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestSubscriptionRepository(t *testing.T) {
	validProduct := &models.Product{
		ID:       uuid.New(),
		Duration: models.DurationMonth, // 30 days
	}
	expiredProduct := &models.Product{
		ID:       uuid.New(),
		Duration: -30, // Already expired duration
	}

	validUserID := uuid.New()

	tests := []struct {
		name          string
		userID        string
		product       *models.Product
		expectError   bool
		expectedError error
	}{
		{
			name:        "Create with valid product",
			userID:      validUserID.String(),
			product:     validProduct,
			expectError: false,
		},
		{
			name:          "Create with nil product",
			userID:        validUserID.String(),
			product:       nil,
			expectError:   true,
			expectedError: repositories.ErrProductRequired,
		},
		{
			name:          "Create with expired product",
			userID:        validUserID.String(),
			product:       expiredProduct,
			expectError:   true,
			expectedError: repositories.ErrInvalidProductDuration,
		},
	}

	repo := repositories.NewSubscriptionRepository()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub, err := repo.CreateSubscription(tt.userID, tt.product)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, sub)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, sub)
				assert.Equal(t, tt.userID, sub.UserID.String())
				assert.Equal(t, tt.product.ID, sub.ProductID)
				assert.Equal(t, models.StatusActive, sub.Status)
				assert.WithinDuration(t, time.Now().Add(time.Duration(tt.product.Duration)*24*time.Hour), sub.EndDate, time.Second)
			}
		})
	}
}

func TestSubscriptionLifecycle(t *testing.T) {
	repo := repositories.NewSubscriptionRepository()
	product := &models.Product{
		ID:       uuid.New(),
		Duration: 30,
	}
	userID := uuid.New()

	// Create subscription
	sub, err := repo.CreateSubscription(userID.String(), product)
	assert.NoError(t, err)
	assert.NotNil(t, sub)

	// Test initial state
	t.Run("Initial state is active", func(t *testing.T) {
		retrieved, err := repo.GetSubscription(sub.ID.String())
		assert.NoError(t, err)
		assert.Equal(t, models.StatusActive, retrieved.Status)
	})

	// Test pause
	t.Run("Can pause active subscription", func(t *testing.T) {
		pausedSub, err := repo.PauseSubscription(sub.ID.String())
		assert.NoError(t, err)
		assert.Equal(t, models.StatusPaused, pausedSub.Status)
	})

	// Test unpause
	t.Run("Can unpause paused subscription", func(t *testing.T) {
		unpausedSub, err := repo.UnpauseSubscription(sub.ID.String())
		assert.NoError(t, err)
		assert.Equal(t, models.StatusActive, unpausedSub.Status)
	})

	// Test cancel
	t.Run("Can cancel active subscription", func(t *testing.T) {
		cancelledSub, err := repo.CancelSubscription(sub.ID.String())
		assert.NoError(t, err)
		assert.Equal(t, models.StatusCancelled, cancelledSub.Status)
	})

	// Test cannot pause cancelled
	t.Run("Cannot pause cancelled subscription", func(t *testing.T) {
		product := &models.Product{ID: uuid.New(), Duration: 30}
		userID := uuid.New()
		sub, err := repo.CreateSubscription(userID.String(), product)
		assert.NoError(t, err)

		_, err = repo.CancelSubscription(sub.ID.String())
		_, err = repo.PauseSubscription(sub.ID.String())

		assert.Error(t, err)
		assert.Equal(t, repositories.ErrCannotPause, err)
	})

	// Verify subscription still exists
	t.Run("Cancelled subscription still exists", func(t *testing.T) {
		product := &models.Product{ID: uuid.New(), Duration: 30}
		userID := uuid.New()
		sub, err := repo.CreateSubscription(userID.String(), product)
		assert.NoError(t, err)
		_, err = repo.CancelSubscription(sub.ID.String())

		retrieved, err := repo.GetSubscription(sub.ID.String())
		assert.NoError(t, err)
		assert.Equal(t, models.StatusCancelled, retrieved.Status)
	})
}

func TestAutoExpiration(t *testing.T) {
	repo := repositories.NewSubscriptionRepository()
	product := &models.Product{
		ID:       uuid.New(),
		Duration: 1, // 1 day duration
	}

	sub, err := repo.CreateSubscription(uuid.New().String(), product)
	assert.NoError(t, err)

	// auto-expire by setting EndDate to past
	sub.EndDate = time.Now().Add(-24 * time.Hour)

	t.Run("Auto-expire on get", func(t *testing.T) {
		retrieved, err := repo.GetSubscription(sub.ID.String())
		assert.NoError(t, err)
		assert.Equal(t, models.StatusExpired, retrieved.Status)
	})
}

func TestUnpauseSubscription(t *testing.T) {
	repo := repositories.NewSubscriptionRepository()
	product := &models.Product{ID: uuid.New(), Duration: 30}
	userID := uuid.New()

	sub, err := repo.CreateSubscription(userID.String(), product)
	assert.NoError(t, err)

	beforePause := time.Now()

	_, err = repo.PauseSubscription(sub.ID.String())
	assert.NoError(t, err)

	pausedSub, err := repo.GetSubscription(sub.ID.String())
	assert.NoError(t, err)
	assert.NotNil(t, pausedSub.PausedAt)
	assert.True(t, pausedSub.PausedAt.After(beforePause) ||
		pausedSub.PausedAt.Equal(beforePause),
		"PausedAt should be after or equal to beforePause time")

	t.Run("Unpause extends subscription correctly", func(t *testing.T) {
		pausedDuration := pausedSub.EndDate.Sub(*pausedSub.PausedAt)

		beforeUnpause := time.Now()

		unpausedSub, err := repo.UnpauseSubscription(sub.ID.String())
		assert.NoError(t, err)

		expectedEnd := beforeUnpause.Add(pausedDuration)
		assert.WithinDuration(t, expectedEnd, unpausedSub.EndDate, time.Second)
		assert.Equal(t, models.StatusActive, unpausedSub.Status)
		assert.Nil(t, unpausedSub.PausedAt)
	})
}
