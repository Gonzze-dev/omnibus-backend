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
	GetByExternalTerminalID(ctx context.Context, externalID uuid.UUID) (models.BusTerminal, error)
	GetByUUIDWithPlatforms(ctx context.Context, id uuid.UUID) (models.BusTerminal, error)
	ListByPostalCode(ctx context.Context, postalCode string) ([]models.BusTerminal, error)
	List(ctx context.Context) ([]models.BusTerminal, error)
	ListPaginated(ctx context.Context, limit, offset int, order string) ([]models.BusTerminal, error)
	Count(ctx context.Context) (int64, error)
	CountActive(ctx context.Context) (int64, error)
	ListWithPlatforms(ctx context.Context) ([]models.BusTerminal, error)
	ListByUUIDs(ctx context.Context, ids []uuid.UUID) ([]models.BusTerminal, error)
	ListWithPlatformsPaginated(ctx context.Context, limit, offset int, order string) ([]models.BusTerminal, error)
	ListByUUIDsPaginated(ctx context.Context, ids []uuid.UUID, limit, offset int, order string) ([]models.BusTerminal, error)
	CountByUUIDs(ctx context.Context, ids []uuid.UUID) (int64, error)
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

func (r *busTerminalRepository) GetByExternalTerminalID(ctx context.Context, externalID uuid.UUID) (models.BusTerminal, error) {
	var bt models.BusTerminal
	err := r.db.WithContext(ctx).Where("external_terminal_id = ?", externalID).First(&bt).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.BusTerminal{}, ErrNotFound
		}
		return models.BusTerminal{}, err
	}
	return bt, nil
}

func (r *busTerminalRepository) GetByUUIDWithPlatforms(ctx context.Context, id uuid.UUID) (models.BusTerminal, error) {
	var bt models.BusTerminal
	err := r.db.WithContext(ctx).
		Where("uuid = ?", id).
		Preload("Platforms", func(db *gorm.DB) *gorm.DB {
			return db.Order("anden")
		}).
		First(&bt).Error
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

func (r *busTerminalRepository) ListPaginated(ctx context.Context, limit, offset int, order string) ([]models.BusTerminal, error) {
	var terminals []models.BusTerminal
	err := r.db.WithContext(ctx).
		Order("name " + order).
		Limit(limit).
		Offset(offset).
		Find(&terminals).Error
	return terminals, err
}

func (r *busTerminalRepository) Count(ctx context.Context) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&models.BusTerminal{}).Count(&total).Error
	return total, err
}

// CountActive cuenta las terminales que tienen al menos una plataforma.
func (r *busTerminalRepository) CountActive(ctx context.Context) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).
		Model(&models.BusTerminal{}).
		Where("EXISTS (SELECT 1 FROM platform p WHERE p.bus_terminal_id = bus_terminal.uuid)").
		Count(&total).Error
	return total, err
}

func (r *busTerminalRepository) ListWithPlatforms(ctx context.Context) ([]models.BusTerminal, error) {
	var terminals []models.BusTerminal
	err := r.db.WithContext(ctx).
		Preload("Platforms", func(db *gorm.DB) *gorm.DB {
			return db.Order("anden")
		}).
		Order("name").
		Find(&terminals).Error
	return terminals, err
}

func (r *busTerminalRepository) ListByUUIDs(ctx context.Context, ids []uuid.UUID) ([]models.BusTerminal, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var terminals []models.BusTerminal
	err := r.db.WithContext(ctx).
		Where("uuid IN ?", ids).
		Preload("Platforms", func(db *gorm.DB) *gorm.DB {
			return db.Order("anden")
		}).
		Order("name").
		Find(&terminals).Error
	return terminals, err
}

func (r *busTerminalRepository) ListWithPlatformsPaginated(ctx context.Context, limit, offset int, order string) ([]models.BusTerminal, error) {
	var terminals []models.BusTerminal
	err := r.db.WithContext(ctx).
		Preload("Platforms", func(db *gorm.DB) *gorm.DB {
			return db.Order("anden")
		}).
		Order("name " + order).
		Limit(limit).
		Offset(offset).
		Find(&terminals).Error
	return terminals, err
}

func (r *busTerminalRepository) ListByUUIDsPaginated(ctx context.Context, ids []uuid.UUID, limit, offset int, order string) ([]models.BusTerminal, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var terminals []models.BusTerminal
	err := r.db.WithContext(ctx).
		Where("uuid IN ?", ids).
		Preload("Platforms", func(db *gorm.DB) *gorm.DB {
			return db.Order("anden")
		}).
		Order("name " + order).
		Limit(limit).
		Offset(offset).
		Find(&terminals).Error
	return terminals, err
}

func (r *busTerminalRepository) CountByUUIDs(ctx context.Context, ids []uuid.UUID) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	var total int64
	err := r.db.WithContext(ctx).
		Model(&models.BusTerminal{}).
		Where("uuid IN ?", ids).
		Count(&total).Error
	return total, err
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
