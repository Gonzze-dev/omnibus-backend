package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"tesina/backend/internal/models"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByUUID(ctx context.Context, id uuid.UUID) (models.User, error)
	GetByEmail(ctx context.Context, email string) (models.User, error)
	ListPaginated(ctx context.Context, search string, limit, offset int, order string) ([]models.User, error)
	Count(ctx context.Context, search string) (int64, error)
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

// searchUsers filtra por nombre completo o email, sin distinguir mayusculas.
// Una busqueda vacia no filtra.
func searchUsers(db *gorm.DB, search string) *gorm.DB {
	search = strings.TrimSpace(search)
	if search == "" {
		return db
	}
	escaped := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(search)
	pattern := "%" + escaped + "%"
	return db.Where(
		"(first_name || ' ' || last_name) ILIKE ? OR email ILIKE ?",
		pattern, pattern,
	)
}

func (r *userRepository) ListPaginated(ctx context.Context, search string, limit, offset int, order string) ([]models.User, error) {
	var users []models.User
	err := searchUsers(r.db.WithContext(ctx), search).
		Preload("Rol").
		Order("first_name " + order).
		Order("last_name " + order).
		Order("uuid").
		Limit(limit).
		Offset(offset).
		Find(&users).Error
	return users, err
}

func (r *userRepository) Count(ctx context.Context, search string) (int64, error) {
	var total int64
	err := searchUsers(r.db.WithContext(ctx).Model(&models.User{}), search).Count(&total).Error
	return total, err
}

func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("uuid = ?", id).Delete(&models.User{}).Error
}
