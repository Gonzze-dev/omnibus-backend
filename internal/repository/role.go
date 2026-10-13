package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"tesina/backend/internal/models"
)

type RoleRepository interface {
	GetByUUID(ctx context.Context, id uuid.UUID) (models.Role, error)
	GetByName(ctx context.Context, name string) (models.Role, error)
	List(ctx context.Context) ([]models.Role, error)
}

type roleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) *roleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) GetByUUID(ctx context.Context, id uuid.UUID) (models.Role, error) {
	var role models.Role
	err := r.db.WithContext(ctx).Where("uuid = ?", id).First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Role{}, ErrNotFound
		}
		return models.Role{}, err
	}
	return role, nil
}

func (r *roleRepository) GetByName(ctx context.Context, name string) (models.Role, error) {
	var role models.Role
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Role{}, ErrNotFound
		}
		return models.Role{}, err
	}
	return role, nil
}

func (r *roleRepository) List(ctx context.Context) ([]models.Role, error) {
	var roles []models.Role
	err := r.db.WithContext(ctx).Order("name").Find(&roles).Error
	return roles, err
}
