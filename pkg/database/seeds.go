package database

import (
	"gymondo_dz/pkg/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func SeedProducts(db *gorm.DB) error {
	var count int64
	db.Model(&models.Product{}).Count(&count)
	if count > 0 {
		return nil
	}

	products := []models.Product{
		{
			Name:        "1-Month Membership",
			Description: "Basic monthly membership",
			Duration:    models.DurationMonth,
			Price:       29.99,
			TaxRate:     0.10,
		},
		{
			Name:        "1-Year Membership",
			Description: "Yearly membership with small discount",
			Duration:    models.DurationYear,
			Price:       79.99,
			TaxRate:     0.10,
		},
		{
			Name:        "Lifetime Membership",
			Description: "Lifetime membership with best discount",
			Duration:    models.DurationLifetime,
			Price:       249.99,
			TaxRate:     0.10,
		},
	}

	return db.Transaction(func(tx *gorm.DB) error {
		for _, p := range products {
			if err := tx.Create(&p).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func SeedSubscriptions(db *gorm.DB) error {
	// First ensure we have products
	var products []models.Product
	if err := db.First(&products).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil // No products exist yet, skip subscription seeding
		}
		return err
	}

	var count int64
	db.Model(&models.Subscription{}).Count(&count)
	if count > 0 {
		return nil
	}

	now := time.Now()
	pausedAt := now.Add(-12 * time.Hour)
	cancelledAt := now.Add(-24 * time.Hour)

	subscriptions := []models.Subscription{
		{
			ID:        uuid.MustParse("AAAAAAAA-AAAA-AAAA-AAAA-AAAAAAAAAAAA"),
			UserID:    uuid.New(),
			ProductID: products[0].ID,
			StartDate: now,
			EndDate:   now.Add(time.Duration(products[0].Duration) * time.Hour),
			Status:    "active",
		},
		{
			ID:        uuid.MustParse("BBBBBBBB-BBBB-BBBB-BBBB-BBBBBBBBBBBB"),
			UserID:    uuid.New(),
			ProductID: products[1%len(products)].ID,
			StartDate: now.Add(-24 * time.Hour), // Started yesterday
			EndDate:   now.Add(time.Duration(products[1%len(products)].Duration) * time.Hour),
			Status:    "paused",
			PausedAt:  &pausedAt,
		},
		{
			ID:          uuid.MustParse("CCCCCCCC-CCCC-CCCC-CCCC-CCCCCCCCCCCC"),
			UserID:      uuid.New(),
			ProductID:   products[2%len(products)].ID,
			StartDate:   now.Add(-7 * 24 * time.Hour), // Started a week ago
			EndDate:     now.Add(time.Duration(products[2%len(products)].Duration) - 7*24*time.Hour),
			Status:      "cancelled",
			CancelledAt: &cancelledAt,
		},
	}

	return db.Transaction(func(tx *gorm.DB) error {
		for _, s := range subscriptions {
			if err := tx.Create(&s).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
