package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SubscriptionDuration int

const (
	DurationMonth    SubscriptionDuration = 30
	DurationYear     SubscriptionDuration = 365
	DurationLifetime SubscriptionDuration = 365 * 100
)

type Product struct {
	ID          uuid.UUID            `gorm:"type:uuid;primaryKey" json:"id"`
	Name        string               `gorm:"size:100;not null" json:"name"`
	Description string               `gorm:"size:255" json:"description,omitempty"`
	Price       float64              `gorm:"type:decimal(10,2);not null" json:"price"`
	TaxRate     float64              `gorm:"type:decimal(5,2);default:0.10" json:"tax_rate"`
	TotalPrice  float64              `gorm:"-" json:"total_price"` // ignored by GORM, only for JSON response
	Duration    SubscriptionDuration `gorm:"not null" json:"duration"`
	CreatedAt   time.Time            `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time            `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt       `gorm:"index" json:"-"` // Explicitly ignored in JSON
}

func (p *Product) BeforeCreate(tx *gorm.DB) (err error) {
	p.ID = uuid.New()
	return
}

func (p *Product) AfterFind(tx *gorm.DB) (err error) {
	p.TotalPrice = p.Price + (p.Price * p.TaxRate)
	return
}
