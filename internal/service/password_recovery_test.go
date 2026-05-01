package service

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"tesina/backend/internal/middleware"
	"tesina/backend/internal/models"
	"tesina/backend/internal/repository"
	"tesina/backend/internal/validators"
	errorsService "tesina/backend/internal/errors"
)

func Test_parsePasswordResetToken_roundTrip(t *testing.T) {
	t.Parallel()
	secret := "reset-secret-test"
	uid := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")

	tok, err := signPasswordResetToken(uid, secret)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	got, exp, err := parsePasswordResetToken(tok, secret)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if got != uid {
		t.Fatalf("user id: got %v want %v", got, uid)
	}
	if !exp.After(time.Now()) {
		t.Fatalf("expected future expiry, got %v", exp)
	}
}

func Test_parsePasswordResetToken_wrongSecret(t *testing.T) {
	t.Parallel()
	uid := uuid.New()
	tok, err := signPasswordResetToken(uid, "secret-a")
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = parsePasswordResetToken(tok, "secret-b")
	if err != errorsService.ErrInvalidPasswordResetToken {
		t.Fatalf("got err %v want errorsService.ErrInvalidPasswordResetToken", err)
	}
}

func Test_parsePasswordResetToken_rejectsSessionJWT(t *testing.T) {
	t.Parallel()
	sessionSecret := "jwt-secret"
	resetSecret := "password-reset-secret"

	claims := middleware.JWTClaims{
		UserID: uuid.New(),
		Role:   "user",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	sessionTok := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims)
	signed, err := sessionTok.SignedString([]byte(sessionSecret))
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = parsePasswordResetToken(signed, resetSecret)
	if err != errorsService.ErrInvalidPasswordResetToken {
		t.Fatalf("got %v want errorsService.ErrInvalidPasswordResetToken", err)
	}
}

func TestForgotPassword_notFound_returnsSameMessage(t *testing.T) {
	t.Parallel()
	svc := NewPasswordRecoveryService(
		stubUserRepo{getByEmailErr: repository.ErrNotFound},
		noopRefreshTokenRepo{},
		nil,
		"reset-secret",
		"https://app.example.com",
		"Test",
	)
	resp, err := svc.ForgotPassword(context.Background(), "nobody@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if resp.Message != forgotPasswordMessage {
		t.Fatalf("message %q", resp.Message)
	}
}

func TestForgotPassword_emptyEmail(t *testing.T) {
	t.Parallel()
	svc := NewPasswordRecoveryService(
		stubUserRepo{},
		noopRefreshTokenRepo{},
		nil,
		"reset-secret",
		"https://app.example.com",
		"Test",
	)
	_, err := svc.ForgotPassword(context.Background(), "   ")
	if err != validators.ErrEmailRequired {
		t.Fatalf("got %v want validators.ErrEmailRequired", err)
	}
}

func TestResetPasswordWithToken_success(t *testing.T) {
	t.Parallel()
	uid := uuid.New()
	var saved *models.User

	repo := stubUserRepo{
		getByUUIDUser: models.User{
			UUID:     uid,
			Email:    "u@example.com",
			Password: "old-hash",
		},
		updateFn: func(u *models.User) error {
			saved = u
			return nil
		},
	}

	svc := NewPasswordRecoveryService(
		repo,
		noopRefreshTokenRepo{},
		nil,
		"reset-secret-xyz",
		"https://app.example.com",
		"Test",
	)
	tok, err := signPasswordResetToken(uid, "reset-secret-xyz")
	if err != nil {
		t.Fatal(err)
	}

	if err := svc.ResetPasswordWithToken(context.Background(), tok, "new-password-9"); err != nil {
		t.Fatal(err)
	}
	if saved == nil {
		t.Fatal("expected user update")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(saved.Password), []byte("new-password-9")); err != nil {
		t.Fatalf("password not updated: %v", err)
	}
}

// --- test doubles ---

type stubUserRepo struct {
	getByEmailUser models.User
	getByEmailErr  error
	getByUUIDUser  models.User
	getByUUIDErr   error
	updateFn       func(*models.User) error
}

func (s stubUserRepo) Create(context.Context, *models.User) error { return nil }

func (s stubUserRepo) GetByUUID(_ context.Context, id uuid.UUID) (models.User, error) {
	if s.getByUUIDErr != nil {
		return models.User{}, s.getByUUIDErr
	}
	if s.getByUUIDUser.UUID != uuid.Nil && s.getByUUIDUser.UUID != id {
		return models.User{}, repository.ErrNotFound
	}
	if s.getByUUIDUser.UUID == uuid.Nil {
		return models.User{}, repository.ErrNotFound
	}
	return s.getByUUIDUser, nil
}

func (s stubUserRepo) GetByEmail(context.Context, string) (models.User, error) {
	if s.getByEmailErr != nil {
		return models.User{}, s.getByEmailErr
	}
	if s.getByEmailUser.Email == "" && s.getByEmailErr == nil {
		return models.User{}, repository.ErrNotFound
	}
	return s.getByEmailUser, nil
}

func (s stubUserRepo) GetByDNI(context.Context, string) (models.User, error) {
	return models.User{}, repository.ErrNotFound
}

func (s stubUserRepo) Update(_ context.Context, user *models.User) error {
	if s.updateFn != nil {
		return s.updateFn(user)
	}
	return nil
}

func (s stubUserRepo) Delete(context.Context, uuid.UUID) error { return nil }

type noopRefreshTokenRepo struct{}

func (noopRefreshTokenRepo) Upsert(context.Context, *models.UserRefreshToken) error { return nil }

func (noopRefreshTokenRepo) GetByUserID(context.Context, uuid.UUID) (models.UserRefreshToken, error) {
	return models.UserRefreshToken{}, repository.ErrNotFound
}

func (noopRefreshTokenRepo) GetByToken(context.Context, string) (models.UserRefreshToken, error) {
	return models.UserRefreshToken{}, repository.ErrNotFound
}

func (noopRefreshTokenRepo) DeleteByUserID(context.Context, uuid.UUID) error { return nil }
