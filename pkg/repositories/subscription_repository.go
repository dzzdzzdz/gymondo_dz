package repositories

import (
	"errors"
	"gymondo_dz/pkg/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
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
	db *gorm.DB
}

func NewSubscriptionRepository(db *gorm.DB) SubscriptionRepository {
	return &SubscriptionRepositoryImpl{db: db}
}

func (r *SubscriptionRepositoryImpl) GetSubscription(id string) (*models.Subscription, error) {
	subID, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrInvalidSubscriptionID
	}

	var subscription models.Subscription
	result := r.db.Preload("Product").First(&subscription, "id = ?", subID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrSubscriptionNotFound
		}
		return nil, result.Error
	}

	// auto-expire if needed
	if subscription.EndDate.Before(time.Now()) && subscription.Status != models.StatusExpired {
		err := r.db.Model(&subscription).Updates(map[string]interface{}{
			"status":     models.StatusExpired,
			"updated_at": time.Now(),
		}).Error
		if err != nil {
			return nil, err
		}
	}

	return &subscription, nil
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

	err = r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(newSub).Error; err != nil {
			return err
		}
		return tx.Preload("Product").First(newSub, "id = ?", newSub.ID).Error
	})

	if err != nil {
		return nil, err
	}

	return newSub, nil
}

func (r *SubscriptionRepositoryImpl) PauseSubscription(id string) (*models.Subscription, error) {
	var subscription models.Subscription
	err := r.db.Transaction(func(tx *gorm.DB) error {
		// Lock the record for update
		if err := tx.Set("gorm:query_option", "FOR UPDATE").
			Preload("Product").
			First(&subscription, "id = ?", id).
			Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrSubscriptionNotFound
			}
			return err
		}

		if subscription.Status != models.StatusActive {
			return ErrCannotPause
		}

		now := time.Now()
		updates := map[string]interface{}{
			"status":     models.StatusPaused,
			"paused_at":  now,
			"updated_at": now,
		}

		return tx.Model(&subscription).Updates(updates).Error
	})

	if err != nil {
		return nil, err
	}

	return &subscription, nil
}

func (r *SubscriptionRepositoryImpl) UnpauseSubscription(id string) (*models.Subscription, error) {
	var subscription models.Subscription
	err := r.db.Transaction(func(tx *gorm.DB) error {
		// Lock the record for update
		if err := tx.Set("gorm:query_option", "FOR UPDATE").
			Preload("Product").
			First(&subscription, "id = ?", id).
			Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrSubscriptionNotFound
			}
			return err
		}

		if subscription.Status != models.StatusPaused {
			return ErrCannotUnpause
		}

		if subscription.PausedAt == nil {
			return errors.New("paused subscription missing PausedAt timestamp")
		}

		remainingDuration := subscription.EndDate.Sub(*subscription.PausedAt)
		now := time.Now()

		updates := map[string]interface{}{
			"status":     models.StatusActive,
			"end_date":   now.Add(remainingDuration),
			"paused_at":  nil,
			"updated_at": now,
		}

		return tx.Model(&subscription).Updates(updates).Error
	})

	if err != nil {
		return nil, err
	}

	return &subscription, nil
}

func (r *SubscriptionRepositoryImpl) CancelSubscription(id string) (*models.Subscription, error) {
	var subscription models.Subscription
	err := r.db.Transaction(func(tx *gorm.DB) error {
		// Lock the record for update
		if err := tx.Set("gorm:query_option", "FOR UPDATE").
			Preload("Product").
			First(&subscription, "id = ?", id).
			Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrSubscriptionNotFound
			}
			return err
		}

		if subscription.Status == models.StatusCancelled {
			return ErrCannotCancel
		}

		now := time.Now()
		updates := map[string]interface{}{
			"status":       models.StatusCancelled,
			"cancelled_at": now,
			"updated_at":   now,
		}

		return tx.Model(&subscription).Updates(updates).Error
	})

	if err != nil {
		return nil, err
	}

	return &subscription, nil
}
