package models

import (
	"time"

	"github.com/google/uuid"
)

type SubscriptionDuration int // make it simple for now

const (
	DurationMonth    SubscriptionDuration = 30
	DurationYear     SubscriptionDuration = 365
	DurationLifetime SubscriptionDuration = 365 * 1000
)

type Product struct {
	ID          uuid.UUID            `json:"id"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Price       float64              `json:"price"`
	Duration    SubscriptionDuration `json:"duration"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
}
