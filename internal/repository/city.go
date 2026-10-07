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
	ListPaginated(ctx context.Context, limit, offset int, order string) ([]models.City, error)
	Count(ctx context.Context) (int64, error)
	CountActive(ctx context.Context) (int64, error)
	Create(ctx context.Context, city *models.City) error
	Update(ctx context.Context, city *models.City) error
	UpdatePostalCode(ctx context.Context, oldPostalCode, newPostalCode string) error
	Delete(ctx context.Context, postalCode string) error
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

func (r *cityRepository) ListPaginated(ctx context.Context, limit, offset int, order string) ([]models.City, error) {
	var cities []models.City
	err := r.db.WithContext(ctx).
		Order("name " + order).
		Limit(limit).
		Offset(offset).
		Find(&cities).Error
	return cities, err
}

func (r *cityRepository) Count(ctx context.Context) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&models.City{}).Count(&total).Error
	return total, err
}

// CountActive cuenta las ciudades que tienen al menos una terminal.
func (r *cityRepository) CountActive(ctx context.Context) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).
		Model(&models.City{}).
		Where("EXISTS (SELECT 1 FROM bus_terminal bt WHERE bt.postal_code = city.postal_code)").
		Count(&total).Error
	return total, err
}

func (r *cityRepository) Create(ctx context.Context, city *models.City) error {
	return r.db.WithContext(ctx).Create(city).Error
}

func (r *cityRepository) Update(ctx context.Context, city *models.City) error {
	return r.db.WithContext(ctx).Save(city).Error
}

func (r *cityRepository) UpdatePostalCode(ctx context.Context, oldPostalCode, newPostalCode string) error {
	return r.db.WithContext(ctx).
		Model(&models.City{}).
		Where("postal_code = ?", oldPostalCode).
		Update("postal_code", newPostalCode).Error
}

func (r *cityRepository) Delete(ctx context.Context, postalCode string) error {
	return r.db.WithContext(ctx).Where("postal_code = ?", postalCode).Delete(&models.City{}).Error
}
