package repository

import (
	"context"

	"gorm.io/gorm"

	"tesina/backend/internal/models"
)

type AwaitedTripRepository interface {
	Save(ctx context.Context, a models.AwaitedTrip) error
	GetEmailsByGroupKey(ctx context.Context, groupKey string) ([]string, error)
	DeleteByGroupKey(ctx context.Context, groupKey string) error
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

func (r *awaitedTripRepository) GetEmailsByGroupKey(ctx context.Context, groupKey string) ([]string, error) {
	var emails []string
	err := r.db.WithContext(ctx).
		Raw("SELECT u.email FROM awaited_trip awt JOIN users u ON awt.user_id = u.uuid WHERE awt.group_key = ?", groupKey).
		Scan(&emails).Error
	return emails, err
}

func (r *awaitedTripRepository) DeleteByGroupKey(ctx context.Context, groupKey string) error {
	return r.db.WithContext(ctx).
		Where("group_key = ?", groupKey).
		Delete(&models.AwaitedTrip{}).Error
}
