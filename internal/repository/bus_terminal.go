package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"tesina/backend/internal/models"
)

type BusTerminalRepository interface {
	GetByUUID(ctx context.Context, id uuid.UUID) (models.BusTerminal, error)
	ListByPostalCode(ctx context.Context, postalCode string) ([]models.BusTerminal, error)
	List(ctx context.Context) ([]models.BusTerminal, error)
	Create(ctx context.Context, terminal *models.BusTerminal) error
	Update(ctx context.Context, terminal *models.BusTerminal) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type busTerminalRepository struct {
	db *gorm.DB
}

func NewBusTerminalRepository(db *gorm.DB) *busTerminalRepository {
	return &busTerminalRepository{db: db}
}

func (r *busTerminalRepository) GetByUUID(ctx context.Context, id uuid.UUID) (models.BusTerminal, error) {
	var bt models.BusTerminal
	err := r.db.WithContext(ctx).Where("uuid = ?", id).First(&bt).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.BusTerminal{}, ErrNotFound
		}
		return models.BusTerminal{}, err
	}
	return bt, nil
}

func (r *busTerminalRepository) ListByPostalCode(ctx context.Context, postalCode string) ([]models.BusTerminal, error) {
	var terminals []models.BusTerminal
	err := r.db.WithContext(ctx).Where("postal_code = ?", postalCode).Order("name").Find(&terminals).Error
	return terminals, err
}

func (r *busTerminalRepository) List(ctx context.Context) ([]models.BusTerminal, error) {
	var terminals []models.BusTerminal
	err := r.db.WithContext(ctx).Order("name").Find(&terminals).Error
	return terminals, err
}

func (r *busTerminalRepository) Create(ctx context.Context, terminal *models.BusTerminal) error {
	return r.db.WithContext(ctx).Create(terminal).Error
}

func (r *busTerminalRepository) Update(ctx context.Context, terminal *models.BusTerminal) error {
	return r.db.WithContext(ctx).Save(terminal).Error
}

func (r *busTerminalRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("uuid = ?", id).Delete(&models.BusTerminal{}).Error
}
