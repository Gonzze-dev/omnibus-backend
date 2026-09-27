package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"tesina/backend/internal/models"
)

type AwaitedTripRepository interface {
	Upsert(ctx context.Context, a models.AwaitedTrip) error
	GetByUserID(ctx context.Context, userID uuid.UUID) (models.AwaitedTrip, error)
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
	MarkNotifiedByGroupKey(ctx context.Context, groupKey string) ([]string, error)
}

type awaitedTripRepository struct {
	db *gorm.DB
}

func NewAwaitedTripRepository(db *gorm.DB) *awaitedTripRepository {
	return &awaitedTripRepository{db: db}
}

// Upsert reemplaza el viaje que el usuario estaba esperando (si había uno).
func (r *awaitedTripRepository) Upsert(ctx context.Context, a models.AwaitedTrip) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}},
			UpdateAll: true,
		}).
		Create(&a).Error
}

func (r *awaitedTripRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (models.AwaitedTrip, error) {
	var a models.AwaitedTrip
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&a).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.AwaitedTrip{}, ErrNotFound
	}
	return a, err
}

func (r *awaitedTripRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&models.AwaitedTrip{}).Error
}

// MarkNotifiedByGroupKey marca como notificados a los pasajeros que esperaban
// ese colectivo y devuelve sus emails. Es atómico: si la cámara detecta el
// mismo colectivo dos veces, la segunda llamada no devuelve a nadie.
func (r *awaitedTripRepository) MarkNotifiedByGroupKey(ctx context.Context, groupKey string) ([]string, error) {
	var emails []string
	err := r.db.WithContext(ctx).
		Raw(`UPDATE awaited_trip awt
			SET notified_at = now()
			FROM users u
			WHERE awt.user_id = u.uuid
			  AND awt.group_key = ?
			  AND awt.notified_at IS NULL
			RETURNING u.email`, groupKey).
		Scan(&emails).Error
	return emails, err
}
