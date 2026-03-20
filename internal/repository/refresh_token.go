package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"tesina/backend/internal/models"
)

type RefreshTokenRepository interface {
	Upsert(ctx context.Context, token *models.UserRefreshToken) error
	GetByUserID(ctx context.Context, userID uuid.UUID) (models.UserRefreshToken, error)
	GetByToken(ctx context.Context, tokenHash string) (models.UserRefreshToken, error)
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
}

type refreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) *refreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) Upsert(ctx context.Context, token *models.UserRefreshToken) error {
	var existing models.UserRefreshToken
	err := r.db.WithContext(ctx).Where("user_id = ?", token.UserID).First(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return r.db.WithContext(ctx).Create(token).Error
		}
		return err
	}
	existing.Token = token.Token
	existing.ExpiryDate = token.ExpiryDate
	return r.db.WithContext(ctx).Save(&existing).Error
}

func (r *refreshTokenRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (models.UserRefreshToken, error) {
	var token models.UserRefreshToken
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&token).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.UserRefreshToken{}, ErrNotFound
		}
		return models.UserRefreshToken{}, err
	}
	return token, nil
}

func (r *refreshTokenRepository) GetByToken(ctx context.Context, tokenHash string) (models.UserRefreshToken, error) {
	var token models.UserRefreshToken
	err := r.db.WithContext(ctx).Where("token = ?", tokenHash).First(&token).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.UserRefreshToken{}, ErrNotFound
		}
		return models.UserRefreshToken{}, err
	}
	return token, nil
}

func (r *refreshTokenRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&models.UserRefreshToken{}).Error
}
