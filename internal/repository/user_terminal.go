package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"tesina/backend/internal/models"
)

type UserTerminalRepository interface {
	Create(ctx context.Context, ut *models.UserTerminal) error
	Delete(ctx context.Context, userID, busTerminalID uuid.UUID) error
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.UserTerminal, error)
	ListByTerminalID(ctx context.Context, busTerminalID uuid.UUID) ([]models.UserTerminal, error)
	Exists(ctx context.Context, userID, busTerminalID uuid.UUID) (bool, error)
}

type userTerminalRepository struct {
	db *gorm.DB
}

func NewUserTerminalRepository(db *gorm.DB) *userTerminalRepository {
	return &userTerminalRepository{db: db}
}

func (r *userTerminalRepository) Create(ctx context.Context, ut *models.UserTerminal) error {
	return r.db.WithContext(ctx).Create(ut).Error
}

func (r *userTerminalRepository) Delete(ctx context.Context, userID, busTerminalID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND bus_terminal_id = ?", userID, busTerminalID).
		Delete(&models.UserTerminal{}).Error
}

func (r *userTerminalRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.UserTerminal, error) {
	var uts []models.UserTerminal
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&uts).Error
	return uts, err
}

func (r *userTerminalRepository) ListByTerminalID(ctx context.Context, busTerminalID uuid.UUID) ([]models.UserTerminal, error) {
	var uts []models.UserTerminal
	err := r.db.WithContext(ctx).Where("bus_terminal_id = ?", busTerminalID).Find(&uts).Error
	return uts, err
}

func (r *userTerminalRepository) Exists(ctx context.Context, userID, busTerminalID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.UserTerminal{}).
		Where("user_id = ? AND bus_terminal_id = ?", userID, busTerminalID).
		Count(&count).Error
	return count > 0, err
}
