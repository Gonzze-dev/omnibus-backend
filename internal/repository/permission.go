package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"tesina/backend/internal/models"
)

type PermissionRepository interface {
	GetByUUID(ctx context.Context, id uuid.UUID) (models.Permission, error)
	ListByRolID(ctx context.Context, rolID uuid.UUID) ([]models.Permission, error)
}

type permissionRepository struct {
	db *gorm.DB
}

func NewPermissionRepository(db *gorm.DB) *permissionRepository {
	return &permissionRepository{db: db}
}

func (r *permissionRepository) GetByUUID(ctx context.Context, id uuid.UUID) (models.Permission, error) {
	var perm models.Permission
	err := r.db.WithContext(ctx).Where("uuid = ?", id).First(&perm).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Permission{}, ErrNotFound
		}
		return models.Permission{}, err
	}
	return perm, nil
}

func (r *permissionRepository) ListByRolID(ctx context.Context, rolID uuid.UUID) ([]models.Permission, error) {
	var perms []models.Permission
	err := r.db.WithContext(ctx).
		Joins("JOIN rol_permissions rp ON rp.permissions_id = permissions.uuid").
		Where("rp.rol_id = ?", rolID).
		Find(&perms).Error
	return perms, err
}
