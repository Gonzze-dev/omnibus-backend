package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"tesina/backend/internal/models"
	"tesina/backend/internal/repository"
	"tesina/backend/internal/validators"
)

type UserService interface {
	GetProfile(ctx context.Context, userID uuid.UUID) (models.UserResponse, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, req models.UpdateUserRequest) (models.UserResponse, error)
	DeleteAccount(ctx context.Context, userID uuid.UUID) error
	ListTerminals(ctx context.Context) ([]models.BusTerminalResponse, error)
}

type userService struct {
	userRepo          repository.UserRepository
	refreshTokenRepo  repository.RefreshTokenRepository
	busTerminalRepo   repository.BusTerminalRepository
	userTerminalRepo  repository.UserTerminalRepository
}

func NewUserService(
	userRepo repository.UserRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	busTerminalRepo repository.BusTerminalRepository,
	userTerminalRepo repository.UserTerminalRepository,
) *userService {
	return &userService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		busTerminalRepo:  busTerminalRepo,
		userTerminalRepo: userTerminalRepo,
	}
}

func (s *userService) GetProfile(ctx context.Context, userID uuid.UUID) (models.UserResponse, error) {
	user, err := s.userRepo.GetByUUID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return models.UserResponse{}, ErrUserNotFound
		}
		return models.UserResponse{}, err
	}
	resp := models.ToUserResponse(user)
	if user.Rol != nil && user.Rol.Name == "admin" {
		uts, err := s.userTerminalRepo.GetByUserID(ctx, userID)
		if err != nil {
			return models.UserResponse{}, err
		}
		if len(uts) == 0 {
			return resp, nil
		}
		ids := make([]uuid.UUID, len(uts))
		for i := range uts {
			ids[i] = uts[i].BusTerminalID
		}
		terminals, err := s.busTerminalRepo.ListByUUIDs(ctx, ids)
		if err != nil {
			return models.UserResponse{}, err
		}
		resp.Terminals = make([]models.ProfileTerminalRef, 0, len(terminals))
		for _, t := range terminals {
			resp.Terminals = append(resp.Terminals, models.ProfileTerminalRef{
				UUID: t.UUID,
				Name: t.Name,
			})
		}
	}
	return resp, nil
}

func (s *userService) UpdateProfile(ctx context.Context, userID uuid.UUID, req models.UpdateUserRequest) (models.UserResponse, error) {
	if err := validators.ValidateUpdateUserRequest(req); err != nil {
		return models.UserResponse{}, err
	}

	user, err := s.userRepo.GetByUUID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return models.UserResponse{}, ErrUserNotFound
		}
		return models.UserResponse{}, err
	}

	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		user.LastName = *req.LastName
	}
	if req.Email != nil {
		existing, err := s.userRepo.GetByEmail(ctx, *req.Email)
		if err == nil && existing.UUID != userID {
			return models.UserResponse{}, ErrEmailAlreadyExists
		}
		user.Email = *req.Email
	}
	if req.DNI != nil {
		existing, err := s.userRepo.GetByDNI(ctx, *req.DNI)
		if err == nil && existing.UUID != userID {
			return models.UserResponse{}, ErrDNIAlreadyExists
		}
		user.DNI = *req.DNI
	}
	if req.Password != nil {
		hashed, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			return models.UserResponse{}, fmt.Errorf("failed to hash password: %w", err)
		}
		user.Password = string(hashed)
	}

	if err := s.userRepo.Update(ctx, &user); err != nil {
		return models.UserResponse{}, fmt.Errorf("failed to update user: %w", err)
	}

	return models.ToUserResponse(user), nil
}

func (s *userService) DeleteAccount(ctx context.Context, userID uuid.UUID) error {
	_, err := s.userRepo.GetByUUID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	_ = s.refreshTokenRepo.DeleteByUserID(ctx, userID)

	if err := s.userRepo.Delete(ctx, userID); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

func (s *userService) ListTerminals(ctx context.Context) ([]models.BusTerminalResponse, error) {
	terminals, err := s.busTerminalRepo.List(ctx)
	if err != nil {
		return nil, err
	}
	return models.ToBusTerminalResponses(terminals), nil
}
