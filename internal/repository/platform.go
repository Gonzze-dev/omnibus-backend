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
	ListByBusTerminalID(ctx context.Context, busTerminalID uuid.UUID) ([]models.Platform, error)
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

func (r *platformRepository) ListByBusTerminalID(ctx context.Context, busTerminalID uuid.UUID) ([]models.Platform, error) {
	var platforms []models.Platform
	err := r.db.WithContext(ctx).
			Where("bus_terminal_id = ?", busTerminalID).
			Order("anden").
			Find(&platforms).
			Error

	return platforms, err
}
