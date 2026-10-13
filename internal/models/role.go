package models

import "github.com/google/uuid"

type Role struct {
	UUID uuid.UUID `json:"uuid" gorm:"primaryKey;column:uuid;type:uuid;default:uuid_generate_v4()"`
	Name string    `json:"name" gorm:"column:name;not null;uniqueIndex"`
}

func (Role) TableName() string {
	return "role"
}
