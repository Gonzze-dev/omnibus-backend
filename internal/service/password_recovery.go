package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"tesina/backend/internal/mail"
	"tesina/backend/internal/models"
	"tesina/backend/internal/repository"
)

const (
	passwordResetTokenDuration = 30 * time.Minute
	forgotPasswordMessage      = "correo enviado"
)

// PasswordResetClaims is the JWT payload for password recovery (signed with PASSWORD_RESET_JWT_SECRET).
type PasswordResetClaims struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}

type PasswordRecoveryService interface {
	ForgotPassword(ctx context.Context, email string) (models.ForgotPasswordResponse, error)
	ValidateRecoveryToken(ctx context.Context, token string) (models.ValidateRecoveryTokenResponse, error)
	ResetPasswordWithToken(ctx context.Context, token, password string) error
}

type passwordRecoveryService struct {
	userRepo         repository.UserRepository
	refreshTokenRepo repository.RefreshTokenRepository
	mailer           *mail.Mailer
	resetJWTSecret   string
	frontEndBaseLink string
	mailSiteName     string
}

func NewPasswordRecoveryService(
	userRepo repository.UserRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	mailer *mail.Mailer,
	resetJWTSecret, frontEndBaseLink, mailSiteName string,
) *passwordRecoveryService {
	return &passwordRecoveryService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		mailer:           mailer,
		resetJWTSecret:   resetJWTSecret,
		frontEndBaseLink: frontEndBaseLink,
		mailSiteName:     mailSiteName,
	}
}

func (s *passwordRecoveryService) ForgotPassword(ctx context.Context, email string) (models.ForgotPasswordResponse, error) {
	email = strings.TrimSpace(email)
	if email == "" {
		return models.ForgotPasswordResponse{}, ErrMissingFields
	}

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return models.ForgotPasswordResponse{Message: forgotPasswordMessage}, nil
		}
		return models.ForgotPasswordResponse{}, err
	}

	if s.resetJWTSecret == "" {
		log.Printf("password recovery: PASSWORD_RESET_JWT_SECRET not set; skipping email for user %s", user.UUID)
		return models.ForgotPasswordResponse{Message: forgotPasswordMessage}, nil
	}

	token, err := signPasswordResetToken(user.UUID, s.resetJWTSecret)
	if err != nil {
		log.Printf("password recovery: sign token: %v", err)
		return models.ForgotPasswordResponse{Message: forgotPasswordMessage}, nil
	}

	base := strings.TrimRight(s.frontEndBaseLink, "/")
	if base == "" {
		log.Printf("password recovery: FRONT_END_BASE_LINK not set; skipping email for user %s", user.UUID)
		return models.ForgotPasswordResponse{Message: forgotPasswordMessage}, nil
	}

	resetLink := base + "/reset-password?token=" + url.QueryEscape(token)

	if s.mailer == nil {
		log.Printf("password recovery: SMTP not configured; skipping email for user %s", user.UUID)
		return models.ForgotPasswordResponse{Message: forgotPasswordMessage}, nil
	}

	htmlBody, err := mail.RenderResetPasswordHTML(mail.ResetPasswordEmailData{
		Link:     resetLink,
		SiteName: s.mailSiteName,
	})
	if err != nil {
		log.Printf("password recovery: render template: %v", err)
		return models.ForgotPasswordResponse{Message: forgotPasswordMessage}, nil
	}

	if err := s.mailer.Send(mail.SendOptions{
		To:      []string{user.Email},
		Subject: "Recuperar contraseña",
		Body:    htmlBody,
		IsHTML:  true,
	}); err != nil {
		log.Printf("password recovery: send mail: %v", err)
	}

	return models.ForgotPasswordResponse{Message: forgotPasswordMessage}, nil
}

func (s *passwordRecoveryService) ValidateRecoveryToken(ctx context.Context, token string) (models.ValidateRecoveryTokenResponse, error) {
	if strings.TrimSpace(token) == "" {
		return models.ValidateRecoveryTokenResponse{}, ErrInvalidPasswordResetToken
	}
	_, exp, err := parsePasswordResetToken(token, s.resetJWTSecret)
	if err != nil {
		return models.ValidateRecoveryTokenResponse{}, err
	}
	expUTC := exp.UTC()
	return models.ValidateRecoveryTokenResponse{Valid: true, ExpiresAt: &expUTC}, nil
}

func (s *passwordRecoveryService) ResetPasswordWithToken(ctx context.Context, token, password string) error {
	if strings.TrimSpace(password) == "" {
		return ErrMissingFields
	}
	userID, _, err := parsePasswordResetToken(token, s.resetJWTSecret)
	if err != nil {
		return err
	}

	user, err := s.userRepo.GetByUUID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrInvalidPasswordResetToken
		}
		return err
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	user.Password = string(hashed)

	if err := s.userRepo.Update(ctx, &user); err != nil {
		return fmt.Errorf("update user: %w", err)
	}

	if err := s.refreshTokenRepo.DeleteByUserID(ctx, userID); err != nil {
		return fmt.Errorf("invalidate refresh sessions: %w", err)
	}

	return nil
}

func signPasswordResetToken(userID uuid.UUID, secret string) (string, error) {
	claims := PasswordResetClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(passwordResetTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims)
	return t.SignedString([]byte(secret))
}

func parsePasswordResetToken(tokenString, secret string) (uuid.UUID, time.Time, error) {
	if secret == "" {
		return uuid.Nil, time.Time{}, ErrInvalidPasswordResetToken
	}
	claims := &PasswordResetClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return uuid.Nil, time.Time{}, ErrInvalidPasswordResetToken
	}
	if claims.UserID == uuid.Nil {
		return uuid.Nil, time.Time{}, ErrInvalidPasswordResetToken
	}
	if claims.ExpiresAt == nil || time.Now().After(claims.ExpiresAt.Time) {
		return uuid.Nil, time.Time{}, ErrInvalidPasswordResetToken
	}
	return claims.UserID, claims.ExpiresAt.Time, nil
}
