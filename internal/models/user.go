package models

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type User struct {
	UUID      uuid.UUID `json:"uuid" gorm:"primaryKey;column:uuid;type:uuid;default:uuid_generate_v4()"`
	FirstName string    `json:"first_name" gorm:"column:first_name;not null"`
	LastName  string    `json:"last_name" gorm:"column:last_name;not null"`
	Email     string    `json:"email" gorm:"column:email;not null;uniqueIndex"`
	Password  string    `json:"-" gorm:"column:password;not null"`
	RolID     uuid.UUID `json:"rol_id" gorm:"column:rol_id;type:uuid;not null"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
	Rol       *Rol      `json:"rol,omitempty" gorm:"foreignKey:RolID;references:UUID"`
}

func (User) TableName() string {
	return "users"
}

type CreateUserRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"-"`
	User         UserResponse `json:"user"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"-"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

type ForgotPasswordResponse struct {
	Message string `json:"message"`
}

type ValidateRecoveryTokenResponse struct {
	Valid     bool       `json:"valid"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

type ResetPasswordRequest struct {
	Password string `json:"password"`
}

type UpdateUserRequest struct {
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	Email     *string `json:"email,omitempty"`
	Password  *string `json:"password,omitempty"`
}

// ProfileTerminalRef is a minimal terminal reference for GET /users/me (admin only).
type ProfileTerminalRef struct {
	UUID uuid.UUID `json:"uuid"`
	Name string    `json:"name"`
}

type UserResponse struct {
	UUID      uuid.UUID              `json:"uuid"`
	FirstName string                 `json:"first_name"`
	LastName  string                 `json:"last_name"`
	Email     string                 `json:"email"`
	Rol       string                 `json:"rol"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	Terminals []ProfileTerminalRef   `json:"terminals,omitempty"`
}

func ToUserResponse(u User) UserResponse {
	rolName := ""
	if u.Rol != nil {
		rolName = u.Rol.Name
	}
	return UserResponse{
		UUID:      u.UUID,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Email:     u.Email,
		Rol:       rolName,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// AdminUserByEmailResponse is the JSON for GET /api/admin/users/by-email.
type AdminUserByEmailResponse struct {
	Name      string   `json:"name"`
	Email     string   `json:"email"`
	Role      string   `json:"role"`
	Terminals []string `json:"terminals,omitempty"`
}

func ToAdminUserByEmailResponse(u User, terminals []string) AdminUserByEmailResponse {
	fullName := strings.TrimSpace(u.FirstName + " " + u.LastName)
	roleName := ""
	if u.Rol != nil {
		roleName = u.Rol.Name
	}
	return AdminUserByEmailResponse{
		Name:      fullName,
		Email:     u.Email,
		Role:      roleName,
		Terminals: terminals,
	}
}
