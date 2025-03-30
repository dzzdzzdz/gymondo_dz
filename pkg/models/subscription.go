package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SubscriptionStatus string

const (
	StatusActive    SubscriptionStatus = "active"
	StatusPaused    SubscriptionStatus = "paused"
	StatusCancelled SubscriptionStatus = "cancelled"
	StatusExpired   SubscriptionStatus = "expired"
)

type Subscription struct {
	ID          uuid.UUID          `gorm:"type:uuid;primaryKey" json:"id"`
	UserID      uuid.UUID          `gorm:"type:uuid;not null" json:"user_id"`
	ProductID   uuid.UUID          `gorm:"type:uuid;not null" json:"product_id"`
	Product     *Product           `gorm:"foreignKey:ProductID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"product,omitempty"`
	StartDate   time.Time          `gorm:"not null" json:"start_date"`
	EndDate     time.Time          `gorm:"not null" json:"end_date"`
	Status      SubscriptionStatus `gorm:"type:varchar(20);not null;default:'active'" json:"status"`
	PausedAt    *time.Time         `gorm:"index" json:"paused_at,omitempty"`
	CancelledAt *time.Time         `gorm:"index" json:"cancelled_at,omitempty"`
	CreatedAt   time.Time          `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time          `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt     `gorm:"index" json:"-"`     // Explicitly ignored in JSON
	Version     int                `gorm:"default:1" json:"-"` // Version for optimistic locking
}

func (s *Subscription) BeforeCreate(tx *gorm.DB) (err error) {
	s.ID = uuid.New()
	return
}
