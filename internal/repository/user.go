package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"tesina/backend/internal/models"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByUUID(ctx context.Context, id uuid.UUID) (models.User, error)
	GetByEmail(ctx context.Context, email string) (models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *userRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) GetByUUID(ctx context.Context, id uuid.UUID) (models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Preload("Rol").Where("uuid = ?", id).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.User{}, ErrNotFound
		}
		return models.User{}, err
	}
	return user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Preload("Rol").Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.User{}, ErrNotFound
		}
		return models.User{}, err
	}
	return user, nil
}

func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("uuid = ?", id).Delete(&models.User{}).Error
}
