package repository

import (
	"context"

	"gorm.io/gorm"

	"tesina/backend/internal/models"
)

type AwaitedTripRepository interface {
	Save(ctx context.Context, a models.AwaitedTrip) error
}

type awaitedTripRepository struct {
	db *gorm.DB
}

func NewAwaitedTripRepository(db *gorm.DB) *awaitedTripRepository {
	return &awaitedTripRepository{db: db}
}

func (r *awaitedTripRepository) Save(ctx context.Context, a models.AwaitedTrip) error {
	return r.db.WithContext(ctx).Save(&a).Error
}
