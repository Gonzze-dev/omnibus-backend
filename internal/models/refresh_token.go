package models

import (
	"time"

	"github.com/google/uuid"
)

type UserRefreshToken struct {
	UUID       uuid.UUID `json:"uuid" gorm:"primaryKey;column:uuid;type:uuid;default:uuid_generate_v4()"`
	UserID     uuid.UUID `json:"user_id" gorm:"column:user_id;type:uuid;not null;uniqueIndex"`
	Token      string    `json:"-" gorm:"column:token;not null"`
	ExpiryDate time.Time `json:"expiry_date" gorm:"column:expiry_date;not null"`
	CreatedAt  time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
}

func (UserRefreshToken) TableName() string {
	return "user_refresh_tokens"
}
