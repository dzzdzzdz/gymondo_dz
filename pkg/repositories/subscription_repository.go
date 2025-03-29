package repositories

import (
	"errors"
	"gymondo_dz/pkg/models"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidSubscriptionID  = errors.New("invalid subscription ID format")
	ErrSubscriptionNotFound   = errors.New("error not found")
	ErrCannotPause            = errors.New("subscription cannot be paused")
	ErrCannotUnpause          = errors.New("subscription cannot be unpaused")
	ErrCannotCancel           = errors.New("subscription cannot be cancelled")
	ErrProductRequired        = errors.New("product reference required")
	ErrInvalidProductDuration = errors.New("product duration must be positive")
)

type SubscriptionRepository interface {
	GetSubscription(id string) (*models.Subscription, error)
	CreateSubscription(id string, product *models.Product) (*models.Subscription, error)
	PauseSubscription(id string) (*models.Subscription, error)
	UnpauseSubscription(id string) (*models.Subscription, error)
	CancelSubscription(id string) (*models.Subscription, error)
}

type SubscriptionRepositoryImpl struct {
	subscriptions map[uuid.UUID]*models.Subscription
}

func NewSubscriptionRepository() SubscriptionRepository {
	return &SubscriptionRepositoryImpl{
		subscriptions: make(map[uuid.UUID]*models.Subscription),
	}
}

func (r *SubscriptionRepositoryImpl) GetSubscription(id string) (*models.Subscription, error) {
	subID, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrInvalidSubscriptionID
	}

	sub, exists := r.subscriptions[subID]
	if !exists {
		return nil, ErrSubscriptionNotFound
	}

	// auto-expire if needed
	if sub.EndDate.Before(time.Now()) && sub.Status != models.StatusExpired {
		sub.Status = models.StatusExpired
		sub.UpdatedAt = time.Now()
	}

	return sub, nil
}

func (r *SubscriptionRepositoryImpl) CreateSubscription(userID string, product *models.Product) (*models.Subscription, error) {
	if product == nil {
		return nil, ErrProductRequired
	}
	if product.Duration <= 0 {
		return nil, ErrInvalidProductDuration
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, ErrInvalidSubscriptionID
	}

	now := time.Now()
	newSub := &models.Subscription{
		ID:        uuid.New(),
		UserID:    userUUID,
		ProductID: product.ID,
		StartDate: now,
		EndDate:   now.Add(time.Hour * 24 * time.Duration(product.Duration)), // Convert days to duration
		Status:    models.StatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}

	r.subscriptions[newSub.ID] = newSub
	return newSub, nil
}

func (r *SubscriptionRepositoryImpl) PauseSubscription(id string) (*models.Subscription, error) {
	sub, err := r.GetSubscription(id)
	if err != nil {
		return nil, err
	}

	if sub.Status != models.StatusActive {
		return nil, ErrCannotPause
	}

	now := time.Now()
	sub.Status = models.StatusPaused
	sub.PausedAt = &now
	sub.UpdatedAt = now

	return sub, nil
}

func (r *SubscriptionRepositoryImpl) UnpauseSubscription(id string) (*models.Subscription, error) {
	sub, err := r.GetSubscription(id)
	if err != nil {
		return nil, err
	}

	if sub.Status != models.StatusPaused {
		return nil, ErrCannotUnpause
	}

	if sub.PausedAt == nil {
		return nil, errors.New("paused subscription missing PausedAt timestamp")
	}
	remainingDuration := sub.EndDate.Sub(*sub.PausedAt)

	now := time.Now()
	sub.Status = models.StatusActive
	sub.EndDate = now.Add(remainingDuration)
	sub.PausedAt = nil
	sub.UpdatedAt = now

	return sub, nil
}

func (r *SubscriptionRepositoryImpl) CancelSubscription(id string) (*models.Subscription, error) {
	sub, err := r.GetSubscription(id)
	if err != nil {
		return nil, err
	}

	if sub.Status == models.StatusCancelled {
		return nil, ErrCannotCancel
	}

	now := time.Now()
	sub.Status = models.StatusCancelled
	sub.CancelledAt = &now
	sub.UpdatedAt = now

	return sub, nil
}
