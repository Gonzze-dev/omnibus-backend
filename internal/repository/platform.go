package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"tesina/backend/internal/models"
)

type PlatformRepository interface {
	GetByCode(ctx context.Context, code int) (models.Platform, error)
	List(ctx context.Context) ([]models.Platform, error)
	ListByBusTerminalID(ctx context.Context, busTerminalID uuid.UUID) ([]models.Platform, error)
	Create(ctx context.Context, platform *models.Platform) error
	Update(ctx context.Context, platform *models.Platform) error
	Delete(ctx context.Context, code int) error
}

type platformRepository struct {
	db *gorm.DB
}

func NewPlatformRepository(db *gorm.DB) *platformRepository {
	return &platformRepository{db: db}
}

func (r *platformRepository) GetByCode(ctx context.Context, code int) (models.Platform, error) {
	var p models.Platform
	err := r.db.WithContext(ctx).
		Where("code = ?", code).
		Preload("BusTerminal").
		First(&p).Error
	
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Platform{}, ErrNotFound
		}
		return models.Platform{}, err
	}
	return p, nil
}

func (r *platformRepository) List(ctx context.Context) ([]models.Platform, error) {
	var platforms []models.Platform
	err := r.db.WithContext(ctx).
		Preload("BusTerminal").
		Order("anden").
		Find(&platforms).
		Error

	return platforms, err
}

func (r *platformRepository) ListByBusTerminalID(ctx context.Context, busTerminalID uuid.UUID) ([]models.Platform, error) {
	var platforms []models.Platform
	err := r.db.WithContext(ctx).
		Where("bus_terminal_id = ?", busTerminalID).
		Order("anden").
		Find(&platforms).
		Error

	return platforms, err
}

func (r *platformRepository) Create(ctx context.Context, platform *models.Platform) error {
	return r.db.WithContext(ctx).Create(platform).Error
}

func (r *platformRepository) Update(ctx context.Context, platform *models.Platform) error {
	return r.db.WithContext(ctx).Save(platform).Error
}

func (r *platformRepository) Delete(ctx context.Context, code int) error {
	return r.db.WithContext(ctx).Where("code = ?", code).Delete(&models.Platform{}).Error
}
