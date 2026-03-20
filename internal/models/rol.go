package models

import "github.com/google/uuid"

type Rol struct {
	UUID uuid.UUID `json:"uuid" gorm:"primaryKey;column:uuid;type:uuid;default:uuid_generate_v4()"`
	Name string    `json:"name" gorm:"column:name;not null;uniqueIndex"`
}

func (Rol) TableName() string {
	return "rol"
}

type Permission struct {
	UUID       uuid.UUID `json:"uuid" gorm:"primaryKey;column:uuid;type:uuid;default:uuid_generate_v4()"`
	NameAction string    `json:"name_action" gorm:"column:name_action;not null;uniqueIndex"`
}

func (Permission) TableName() string {
	return "permissions"
}

type RolPermission struct {
	RolID        uuid.UUID `json:"rol_id" gorm:"primaryKey;column:rol_id;type:uuid"`
	PermissionID uuid.UUID `json:"permissions_id" gorm:"primaryKey;column:permissions_id;type:uuid"`
}

func (RolPermission) TableName() string {
	return "rol_permissions"
}
