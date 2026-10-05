package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"tesina/backend/internal/models"
)

type NotificationRepository interface {
	Insert(ctx context.Context, n models.Notification) error
	List(ctx context.Context) ([]models.Notification, error)
	ListWithFilters(ctx context.Context, f models.NotificationFilters) ([]models.Notification, int64, error)
	ListAdminPaginated(ctx context.Context, f models.AdminNotificationFilters, limit, offset int, order string) ([]models.Notification, error)
	CountAdmin(ctx context.Context, f models.AdminNotificationFilters) (int64, error)
	GetByID(ctx context.Context, id uuid.UUID) (models.Notification, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) *notificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Insert(ctx context.Context, n models.Notification) error {
	return r.db.WithContext(ctx).Create(&n).Error
}

func (r *notificationRepository) List(ctx context.Context) ([]models.Notification, error) {
	var notifications []models.Notification
	err := r.db.WithContext(ctx).Order("date desc").Find(&notifications).Error
	return notifications, err
}

func (r *notificationRepository) ListWithFilters(ctx context.Context, f models.NotificationFilters) ([]models.Notification, int64, error) {
	q := r.db.WithContext(ctx).Model(&models.Notification{})

	// group_key OR block
	var clauses []string
	var args []any
	if f.GroupKeyIsNull {
		clauses = append(clauses, "group_key IS NULL")
	}
	if len(f.GroupKeyExact) > 0 {
		clauses = append(clauses, "group_key IN ?")
		args = append(args, f.GroupKeyExact)
	}
	for _, pattern := range f.GroupKeyLike {
		clauses = append(clauses, "group_key LIKE ?")
		args = append(args, pattern)
	}
	if len(clauses) > 0 {
		q = q.Where("("+strings.Join(clauses, " OR ")+")", args...)
	}

	if f.ExcludeAdminGroups {
		q = q.Where("group_name NOT ILIKE ?", "%admin%")
	}

	if f.NotificationType != nil {
		q = q.Where("payload->>'type' = ?", string(*f.NotificationType))
	}

	if f.OnlyExpired != nil && *f.OnlyExpired {
		q = q.Where("expiration <= NOW()")
	} else {
		q = q.Where("expiration > NOW()")
	}

	if f.StartDate != nil {
		q = q.Where("date >= ?", *f.StartDate)
		if f.EndDate != nil {
			q = q.Where("date <= ?", *f.EndDate)
		}
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var results []models.Notification
	err := q.Order("date desc").Limit(f.Limit).Offset(f.Offset).Find(&results).Error
	return results, total, err
}

// adminNotificationScope aplica los filtros del listado de admin.
func adminNotificationScope(db *gorm.DB, f models.AdminNotificationFilters) *gorm.DB {
	if f.TerminalIDs != nil {
		if len(f.TerminalIDs) == 0 {
			return db.Where("1 = 0")
		}
		clauses := []string{"group_key IN ?"}
		args := []any{f.TerminalIDs}
		for _, tid := range f.TerminalIDs {
			clauses = append(clauses, "group_key LIKE ?")
			args = append(args, "%:"+tid)
		}
		db = db.Where("("+strings.Join(clauses, " OR ")+")", args...)
	}
	if f.Type != nil {
		db = db.Where("payload->>'type' = ?", string(*f.Type))
	}
	switch f.Status {
	case "active":
		db = db.Where("expiration > NOW()")
	case "expired":
		db = db.Where("expiration <= NOW()")
	}
	return db
}

func (r *notificationRepository) ListAdminPaginated(ctx context.Context, f models.AdminNotificationFilters, limit, offset int, order string) ([]models.Notification, error) {
	var results []models.Notification
	err := adminNotificationScope(r.db.WithContext(ctx), f).
		Order("date " + order).
		Order("id " + order).
		Limit(limit).
		Offset(offset).
		Find(&results).Error
	return results, err
}

func (r *notificationRepository) CountAdmin(ctx context.Context, f models.AdminNotificationFilters) (int64, error) {
	var total int64
	err := adminNotificationScope(r.db.WithContext(ctx).Model(&models.Notification{}), f).Count(&total).Error
	return total, err
}

func (r *notificationRepository) GetByID(ctx context.Context, id uuid.UUID) (models.Notification, error) {
	var n models.Notification
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&n).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.Notification{}, ErrNotFound
	}
	return n, err
}

func (r *notificationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&models.Notification{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
