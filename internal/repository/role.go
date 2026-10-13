package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"tesina/backend/internal/models"
)

type RolRepository interface {
	GetByUUID(ctx context.Context, id uuid.UUID) (models.Rol, error)
	GetByName(ctx context.Context, name string) (models.Rol, error)
	List(ctx context.Context) ([]models.Rol, error)
}

type rolRepository struct {
	db *gorm.DB
}

func NewRolRepository(db *gorm.DB) *rolRepository {
	return &rolRepository{db: db}
}

func (r *rolRepository) GetByUUID(ctx context.Context, id uuid.UUID) (models.Rol, error) {
	var rol models.Rol
	err := r.db.WithContext(ctx).Where("uuid = ?", id).First(&rol).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Rol{}, ErrNotFound
		}
		return models.Rol{}, err
	}
	return rol, nil
}

func (r *rolRepository) GetByName(ctx context.Context, name string) (models.Rol, error) {
	var rol models.Rol
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&rol).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Rol{}, ErrNotFound
		}
		return models.Rol{}, err
	}
	return rol, nil
}

func (r *rolRepository) List(ctx context.Context) ([]models.Rol, error) {
	var roles []models.Rol
	err := r.db.WithContext(ctx).Order("name").Find(&roles).Error
	return roles, err
}
