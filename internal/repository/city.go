package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"tesina/backend/internal/models"
)

type CityRepository interface {
	GetByPostalCode(ctx context.Context, postalCode string) (models.City, error)
	List(ctx context.Context) ([]models.City, error)
}

type cityRepository struct {
	db *gorm.DB
}

func NewCityRepository(db *gorm.DB) *cityRepository {
	return &cityRepository{db: db}
}

func (r *cityRepository) GetByPostalCode(ctx context.Context, postalCode string) (models.City, error) {
	var city models.City
	err := r.db.WithContext(ctx).Where("postal_code = ?", postalCode).First(&city).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.City{}, ErrNotFound
		}
		return models.City{}, err
	}
	return city, nil
}

func (r *cityRepository) List(ctx context.Context) ([]models.City, error) {
	var cities []models.City
	err := r.db.WithContext(ctx).Order("name").Find(&cities).Error
	return cities, err
}
