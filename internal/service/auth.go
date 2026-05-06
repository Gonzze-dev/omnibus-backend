package service

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
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

const (
	accessTokenDuration  = 15 * time.Minute
	refreshTokenDuration = 7 * 24 * time.Hour
)

type AuthService interface {
	Register(ctx context.Context, req models.CreateUserRequest) (models.UserResponse, error)
	Login(ctx context.Context, req models.LoginRequest) (models.LoginResponse, error)
	RefreshToken(ctx context.Context, req models.RefreshTokenRequest) (models.RefreshTokenResponse, error)
	Logout(ctx context.Context, req models.LogoutRequest) error
}

type authService struct {
	userRepo         repository.UserRepository
	rolRepo          repository.RolRepository
	refreshTokenRepo repository.RefreshTokenRepository
	jwtSecret        string
}

func NewAuthService(
	userRepo repository.UserRepository,
	rolRepo repository.RolRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	jwtSecret string,
) *authService {
	return &authService{
		userRepo:         userRepo,
		rolRepo:          rolRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtSecret:        jwtSecret,
	}
}

func (s *authService) Register(ctx context.Context, req models.CreateUserRequest) (models.UserResponse, error) {
	if err := validators.ValidateCreateUserRequest(req); err != nil {
		return models.UserResponse{}, err
	}

	if _, err := s.userRepo.GetByEmail(ctx, req.Email); err == nil {
		return models.UserResponse{}, errorsService.ErrEmailAlreadyExists
	}

	rol, err := s.rolRepo.GetByName(ctx, "user")
	if err != nil {
		return models.UserResponse{}, fmt.Errorf("%w: %w", errorsService.ErrRolNotFound, err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return models.UserResponse{}, fmt.Errorf("failed to hash password: %w", err)
	}

	user := models.User{
		UUID:      uuid.New(),
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Password:  string(hashedPassword),
		RolID:     rol.UUID,
	}

	if err := s.userRepo.Create(ctx, &user); err != nil {
		return models.UserResponse{}, fmt.Errorf("failed to create user: %w", err)
	}

	user.Rol = &rol

	return models.ToUserResponse(user), nil
}

func (s *authService) Login(ctx context.Context, req models.LoginRequest) (models.LoginResponse, error) {
	if err := validators.ValidateLoginRequest(req); err != nil {
		return models.LoginResponse{}, err
	}

	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return models.LoginResponse{}, errorsService.ErrInvalidCredentials
		}
		return models.LoginResponse{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return models.LoginResponse{}, errorsService.ErrInvalidCredentials
	}

	roleName := ""
	if user.Rol != nil {
		roleName = user.Rol.Name
	}

	accessToken, err := s.generateAccessToken(user.UUID, roleName)
	if err != nil {
		return models.LoginResponse{}, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken := uuid.New().String()
	hashedRefresh := hashToken(refreshToken)

	rt := models.UserRefreshToken{
		UUID:       uuid.New(),
		UserID:     user.UUID,
		Token:      hashedRefresh,
		ExpiryDate: time.Now().Add(refreshTokenDuration),
	}

	if err := s.refreshTokenRepo.Upsert(ctx, &rt); err != nil {
		return models.LoginResponse{}, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return models.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         models.ToUserResponse(user),
	}, nil
}

func (s *authService) RefreshToken(ctx context.Context, req models.RefreshTokenRequest) (models.RefreshTokenResponse, error) {
	if err := validators.ValidateRefreshTokenRequest(req); err != nil {
		return models.RefreshTokenResponse{}, err
	}

	hashedIncoming := hashToken(req.RefreshToken)

	stored, err := s.refreshTokenRepo.GetByToken(ctx, hashedIncoming)
	if err != nil {
		return models.RefreshTokenResponse{}, errorsService.ErrInvalidRefreshToken
	}

	if time.Now().After(stored.ExpiryDate) {
		_ = s.refreshTokenRepo.DeleteByUserID(ctx, stored.UserID)
		return models.RefreshTokenResponse{}, errorsService.ErrInvalidRefreshToken
	}

	user, err := s.userRepo.GetByUUID(ctx, stored.UserID)
	if err != nil {
		return models.RefreshTokenResponse{}, errorsService.ErrUserNotFound
	}

	roleName := ""
	if user.Rol != nil {
		roleName = user.Rol.Name
	}

	newAccessToken, err := s.generateAccessToken(user.UUID, roleName)
	if err != nil {
		return models.RefreshTokenResponse{}, fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken := uuid.New().String()
	hashedNew := hashToken(newRefreshToken)

	rt := models.UserRefreshToken{
		UUID:       uuid.New(),
		UserID:     user.UUID,
		Token:      hashedNew,
		ExpiryDate: time.Now().Add(refreshTokenDuration),
	}

	if err := s.refreshTokenRepo.Upsert(ctx, &rt); err != nil {
		return models.RefreshTokenResponse{}, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return models.RefreshTokenResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (s *authService) Logout(ctx context.Context, req models.LogoutRequest) error {
	if err := validators.ValidateLogoutRequest(req); err != nil {
		return err
	}

	hashedIncoming := hashToken(req.RefreshToken)

	stored, err := s.refreshTokenRepo.GetByToken(ctx, hashedIncoming)
	if err != nil {
		return errorsService.ErrInvalidRefreshToken
	}

	return s.refreshTokenRepo.DeleteByUserID(ctx, stored.UserID)
}

func (s *authService) generateAccessToken(userID uuid.UUID, role string) (string, error) {
	claims := middleware.JWTClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", h)
}
