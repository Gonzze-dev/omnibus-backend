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
	ListTerminalRefsByUserIDs(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID][]models.ProfileTerminalRef, error)
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

// ListTerminalRefsByUserIDs devuelve, por usuario, las terminales que tiene asignadas (uuid y nombre).
func (r *userTerminalRepository) ListTerminalRefsByUserIDs(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID][]models.ProfileTerminalRef, error) {
	result := make(map[uuid.UUID][]models.ProfileTerminalRef, len(userIDs))
	if len(userIDs) == 0 {
		return result, nil
	}

	var rows []struct {
		UserID uuid.UUID
		UUID   uuid.UUID
		Name   string
	}
	err := r.db.WithContext(ctx).
		Table("user_terminal AS ut").
		Select("ut.user_id, bt.uuid, bt.name").
		Joins("JOIN bus_terminal AS bt ON bt.uuid = ut.bus_terminal_id").
		Where("ut.user_id IN ?", userIDs).
		Order("bt.name").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		result[row.UserID] = append(result[row.UserID], models.ProfileTerminalRef{UUID: row.UUID, Name: row.Name})
	}
	return result, nil
}
