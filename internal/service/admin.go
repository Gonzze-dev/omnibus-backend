package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"tesina/backend/internal/models"
	"tesina/backend/internal/repository"
)

type AdminService interface {
	// Cities
	ListCities(ctx context.Context) ([]models.City, error)
	GetCity(ctx context.Context, postalCode string) (models.City, error)
	CreateCity(ctx context.Context, req models.CreateCityRequest) (models.City, error)
	UpdateCity(ctx context.Context, postalCode string, req models.UpdateCityRequest) (models.City, error)
	DeleteCity(ctx context.Context, postalCode string) error

	// Platforms (only admin's terminals)
	ListAllPlatforms(ctx context.Context, busTerminalID *uuid.UUID) ([]models.BusTerminalWithPlatformsResponse, error)
	ListPlatforms(ctx context.Context, adminID uuid.UUID, busTerminalID *uuid.UUID) ([]models.BusTerminalWithPlatformsResponse, error)
	GetPlatformByCode(ctx context.Context, code int) (models.Platform, error)
	GetPlatform(ctx context.Context, adminID uuid.UUID, code int) (models.Platform, error)
	CreatePlatformDirect(ctx context.Context, req models.CreatePlatformRequest) (models.Platform, error)
	CreatePlatform(ctx context.Context, adminID uuid.UUID, req models.CreatePlatformRequest) (models.Platform, error)
	UpdatePlatformByCode(ctx context.Context, code int, req models.UpdatePlatformRequest) (models.Platform, error)
	UpdatePlatform(ctx context.Context, adminID uuid.UUID, code int, req models.UpdatePlatformRequest) (models.Platform, error)
	DeletePlatformByCode(ctx context.Context, code int) error
	DeletePlatform(ctx context.Context, adminID uuid.UUID, code int) error

	// User management
	PromoteToAdminDirect(ctx context.Context, req models.PromoteAdminRequest) (models.UserResponse, error)
	PromoteToAdmin(ctx context.Context, adminID uuid.UUID, req models.PromoteAdminRequest) (models.UserResponse, error)
	DemoteAdminDirect(ctx context.Context, req models.DemoteAdminRequest) (models.UserResponse, error)
	DemoteAdmin(ctx context.Context, adminID uuid.UUID, req models.DemoteAdminRequest) (models.UserResponse, error)
}

type adminService struct {
	cityRepo         repository.CityRepository
	platformRepo     repository.PlatformRepository
	busTerminalRepo  repository.BusTerminalRepository
	userRepo         repository.UserRepository
	rolRepo          repository.RolRepository
	userTerminalRepo repository.UserTerminalRepository
}

func NewAdminService(
	cityRepo repository.CityRepository,
	platformRepo repository.PlatformRepository,
	busTerminalRepo repository.BusTerminalRepository,
	userRepo repository.UserRepository,
	rolRepo repository.RolRepository,
	userTerminalRepo repository.UserTerminalRepository,
) *adminService {
	return &adminService{
		cityRepo:         cityRepo,
		platformRepo:     platformRepo,
		busTerminalRepo:  busTerminalRepo,
		userRepo:         userRepo,
		rolRepo:          rolRepo,
		userTerminalRepo: userTerminalRepo,
	}
}

// --- Cities ---

func (s *adminService) ListCities(ctx context.Context) ([]models.City, error) {
	return s.cityRepo.List(ctx)
}

func (s *adminService) GetCity(ctx context.Context, postalCode string) (models.City, error) {
	city, err := s.cityRepo.GetByPostalCode(ctx, postalCode)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return models.City{}, ErrCityNotFound
		}
		return models.City{}, err
	}
	return city, nil
}

func (s *adminService) CreateCity(ctx context.Context, req models.CreateCityRequest) (models.City, error) {
	if req.PostalCode == "" || req.Name == "" {
		return models.City{}, ErrMissingFields
	}

	if _, err := s.cityRepo.GetByPostalCode(ctx, req.PostalCode); err == nil {
		return models.City{}, ErrCityAlreadyExists
	}

	city := models.City{
		PostalCode: req.PostalCode,
		Name:       req.Name,
	}
	if err := s.cityRepo.Create(ctx, &city); err != nil {
		return models.City{}, fmt.Errorf("failed to create city: %w", err)
	}
	return city, nil
}

func (s *adminService) UpdateCity(ctx context.Context, postalCode string, req models.UpdateCityRequest) (models.City, error) {
	city, err := s.cityRepo.GetByPostalCode(ctx, postalCode)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return models.City{}, ErrCityNotFound
		}
		return models.City{}, err
	}

	if req.Name != "" {
		city.Name = req.Name
	}

	if err := s.cityRepo.Update(ctx, &city); err != nil {
		return models.City{}, fmt.Errorf("failed to update city: %w", err)
	}
	return city, nil
}

func (s *adminService) DeleteCity(ctx context.Context, postalCode string) error {
	if _, err := s.cityRepo.GetByPostalCode(ctx, postalCode); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrCityNotFound
		}
		return err
	}
	return s.cityRepo.Delete(ctx, postalCode)
}

// --- Platforms (admin's terminals only) ---

func (s *adminService) verifyTerminalOwnership(ctx context.Context, adminID, busTerminalID uuid.UUID) error {
	exists, err := s.userTerminalRepo.Exists(ctx, adminID, busTerminalID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrTerminalNotOwned
	}
	return nil
}

func (s *adminService) ListAllPlatforms(ctx context.Context, busTerminalID *uuid.UUID) ([]models.BusTerminalWithPlatformsResponse, error) {
	if busTerminalID != nil {
		bt, err := s.busTerminalRepo.GetByUUIDWithPlatforms(ctx, *busTerminalID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return nil, ErrTerminalNotFound
			}
			return nil, err
		}
		return models.ToBusTerminalWithPlatformsResponse([]models.BusTerminal{bt}), nil
	}

	terminals, err := s.busTerminalRepo.ListWithPlatforms(ctx)
	if err != nil {
		return nil, err
	}
	return models.ToBusTerminalWithPlatformsResponse(terminals), nil
}

func (s *adminService) ListPlatforms(ctx context.Context, adminID uuid.UUID, busTerminalID *uuid.UUID) ([]models.BusTerminalWithPlatformsResponse, error) {
	if busTerminalID != nil {
		if err := s.verifyTerminalOwnership(ctx, adminID, *busTerminalID); err != nil {
			return nil, err
		}

		bt, err := s.busTerminalRepo.GetByUUIDWithPlatforms(ctx, *busTerminalID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return nil, ErrTerminalNotFound
			}
			return nil, err
		}
		return models.ToBusTerminalWithPlatformsResponse([]models.BusTerminal{bt}), nil
	}

	uts, err := s.userTerminalRepo.GetByUserID(ctx, adminID)
	if err != nil {
		return nil, err
	}

	ids := make([]uuid.UUID, len(uts))
	for i, ut := range uts {
		ids[i] = ut.BusTerminalID
	}

	terminals, err := s.busTerminalRepo.ListByUUIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	return models.ToBusTerminalWithPlatformsResponse(terminals), nil
}

func (s *adminService) GetPlatformByCode(ctx context.Context, code int) (models.Platform, error) {
	platform, err := s.platformRepo.GetByCode(ctx, code)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return models.Platform{}, ErrPlatformNotFound
		}
		return models.Platform{}, err
	}
	return platform, nil
}

func (s *adminService) GetPlatform(ctx context.Context, adminID uuid.UUID, code int) (models.Platform, error) {
	platform, err := s.GetPlatformByCode(ctx, code)
	
	if err != nil {
		return models.Platform{}, err
	}

	if err := s.verifyTerminalOwnership(ctx, adminID, platform.BusTerminalID); err != nil {
		return models.Platform{}, err
	}
	return platform, nil
}

func (s *adminService) CreatePlatformDirect(ctx context.Context, req models.CreatePlatformRequest) (models.Platform, error) {
	if req.Anden == "" {
		return models.Platform{}, ErrMissingFields
	}

	if _, err := s.busTerminalRepo.GetByUUID(ctx, req.BusTerminalID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return models.Platform{}, ErrTerminalNotFound
		}
		return models.Platform{}, err
	}

	platform := models.Platform{
		Anden:         req.Anden,
		Coordinates:   req.Coordinates,
		BusTerminalID: req.BusTerminalID,
	}
	if err := s.platformRepo.Create(ctx, &platform); err != nil {
		return models.Platform{}, fmt.Errorf("failed to create platform: %w", err)
	}
	return platform, nil
}

func (s *adminService) CreatePlatform(ctx context.Context, adminID uuid.UUID, req models.CreatePlatformRequest) (models.Platform, error) {
	if req.Anden == "" {
		return models.Platform{}, ErrMissingFields
	}

	if err := s.verifyTerminalOwnership(ctx, adminID, req.BusTerminalID); err != nil {
		return models.Platform{}, err
	}

	platform := models.Platform{
		Anden:         req.Anden,
		Coordinates:   req.Coordinates,
		BusTerminalID: req.BusTerminalID,
	}
	if err := s.platformRepo.Create(ctx, &platform); err != nil {
		return models.Platform{}, fmt.Errorf("failed to create platform: %w", err)
	}
	return platform, nil
}

func (s *adminService) UpdatePlatformByCode(ctx context.Context, code int, req models.UpdatePlatformRequest) (models.Platform, error) {
	platform, err := s.platformRepo.GetByCode(ctx, code)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return models.Platform{}, ErrPlatformNotFound
		}
		return models.Platform{}, err
	}

	if req.Anden != nil {
		platform.Anden = *req.Anden
	}
	if req.Coordinates != nil {
		platform.Coordinates = *req.Coordinates
	}

	if err := s.platformRepo.Update(ctx, &platform); err != nil {
		return models.Platform{}, fmt.Errorf("failed to update platform: %w", err)
	}
	return platform, nil
}

func (s *adminService) UpdatePlatform(ctx context.Context, adminID uuid.UUID, code int, req models.UpdatePlatformRequest) (models.Platform, error) {
	platform, err := s.platformRepo.GetByCode(ctx, code)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return models.Platform{}, ErrPlatformNotFound
		}
		return models.Platform{}, err
	}

	if err := s.verifyTerminalOwnership(ctx, adminID, platform.BusTerminalID); err != nil {
		return models.Platform{}, err
	}

	if req.Anden != nil {
		platform.Anden = *req.Anden
	}
	if req.Coordinates != nil {
		platform.Coordinates = *req.Coordinates
	}

	if err := s.platformRepo.Update(ctx, &platform); err != nil {
		return models.Platform{}, fmt.Errorf("failed to update platform: %w", err)
	}
	return platform, nil
}

func (s *adminService) DeletePlatformByCode(ctx context.Context, code int) error {
	if _, err := s.platformRepo.GetByCode(ctx, code); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrPlatformNotFound
		}
		return err
	}
	return s.platformRepo.Delete(ctx, code)
}

func (s *adminService) DeletePlatform(ctx context.Context, adminID uuid.UUID, code int) error {
	platform, err := s.platformRepo.GetByCode(ctx, code)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrPlatformNotFound
		}
		return err
	}

	if err := s.verifyTerminalOwnership(ctx, adminID, platform.BusTerminalID); err != nil {
		return err
	}

	return s.platformRepo.Delete(ctx, code)
}

// --- User management ---

func (s *adminService) PromoteToAdminDirect(ctx context.Context, req models.PromoteAdminRequest) (models.UserResponse, error) {
	if req.Email == "" {
		return models.UserResponse{}, ErrMissingFields
	}

	if _, err := s.busTerminalRepo.GetByUUID(ctx, req.BusTerminalID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return models.UserResponse{}, ErrTerminalNotFound
		}
		return models.UserResponse{}, err
	}

	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return models.UserResponse{}, ErrUserNotFound
		}
		return models.UserResponse{}, err
	}

	if user.Rol != nil && user.Rol.Name == "super_admin" {
		return models.UserResponse{}, ErrAlreadySuperAdmin
	}

	adminRol, err := s.rolRepo.GetByName(ctx, "admin")
	if err != nil {
		return models.UserResponse{}, fmt.Errorf("%w: %w", ErrRolNotFound, err)
	}

	user.RolID = adminRol.UUID
	user.Rol = &adminRol
	if err := s.userRepo.Update(ctx, &user); err != nil {
		return models.UserResponse{}, fmt.Errorf("failed to update user role: %w", err)
	}

	exists, _ := s.userTerminalRepo.Exists(ctx, user.UUID, req.BusTerminalID)
	if !exists {
		ut := models.UserTerminal{
			UserID:        user.UUID,
			BusTerminalID: req.BusTerminalID,
		}
		if err := s.userTerminalRepo.Create(ctx, &ut); err != nil {
			return models.UserResponse{}, fmt.Errorf("failed to associate terminal: %w", err)
		}
	}

	return models.ToUserResponse(user), nil
}

func (s *adminService) PromoteToAdmin(ctx context.Context, adminID uuid.UUID, req models.PromoteAdminRequest) (models.UserResponse, error) {
	if req.Email == "" {
		return models.UserResponse{}, ErrMissingFields
	}

	if err := s.verifyTerminalOwnership(ctx, adminID, req.BusTerminalID); err != nil {
		return models.UserResponse{}, err
	}

	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return models.UserResponse{}, ErrUserNotFound
		}
		return models.UserResponse{}, err
	}

	adminRol, err := s.rolRepo.GetByName(ctx, "admin")
	if err != nil {
		return models.UserResponse{}, fmt.Errorf("%w: %w", ErrRolNotFound, err)
	}

	if user.Rol != nil && user.Rol.Name == "super_admin" {
		return models.UserResponse{}, ErrAlreadySuperAdmin
	}

	user.RolID = adminRol.UUID
	user.Rol = &adminRol
	if err := s.userRepo.Update(ctx, &user); err != nil {
		return models.UserResponse{}, fmt.Errorf("failed to update user role: %w", err)
	}

	exists, _ := s.userTerminalRepo.Exists(ctx, user.UUID, req.BusTerminalID)
	if !exists {
		ut := models.UserTerminal{
			UserID:        user.UUID,
			BusTerminalID: req.BusTerminalID,
		}
		if err := s.userTerminalRepo.Create(ctx, &ut); err != nil {
			return models.UserResponse{}, fmt.Errorf("failed to associate terminal: %w", err)
		}
	}

	return models.ToUserResponse(user), nil
}

func (s *adminService) DemoteAdminDirect(ctx context.Context, req models.DemoteAdminRequest) (models.UserResponse, error) {
	if req.Email == "" {
		return models.UserResponse{}, ErrMissingFields
	}

	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return models.UserResponse{}, ErrUserNotFound
		}
		return models.UserResponse{}, err
	}

	if user.Rol == nil || user.Rol.Name != "admin" {
		return models.UserResponse{}, ErrNotAdmin
	}

	if err := s.userTerminalRepo.Delete(ctx, user.UUID, req.BusTerminalID); err != nil {
		return models.UserResponse{}, fmt.Errorf("failed to remove terminal association: %w", err)
	}

	remaining, err := s.userTerminalRepo.GetByUserID(ctx, user.UUID)
	if err != nil {
		return models.UserResponse{}, err
	}

	if len(remaining) == 0 {
		userRol, err := s.rolRepo.GetByName(ctx, "user")
		if err != nil {
			return models.UserResponse{}, fmt.Errorf("%w: %w", ErrRolNotFound, err)
		}
		user.RolID = userRol.UUID
		user.Rol = &userRol
		if err := s.userRepo.Update(ctx, &user); err != nil {
			return models.UserResponse{}, fmt.Errorf("failed to update user role: %w", err)
		}
	}

	return models.ToUserResponse(user), nil
}

func (s *adminService) DemoteAdmin(ctx context.Context, adminID uuid.UUID, req models.DemoteAdminRequest) (models.UserResponse, error) {
	if req.Email == "" {
		return models.UserResponse{}, ErrMissingFields
	}

	if err := s.verifyTerminalOwnership(ctx, adminID, req.BusTerminalID); err != nil {
		return models.UserResponse{}, err
	}

	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return models.UserResponse{}, ErrUserNotFound
		}
		return models.UserResponse{}, err
	}

	if user.UUID == adminID {
		return models.UserResponse{}, ErrCannotDemoteSelf
	}

	if user.Rol == nil || user.Rol.Name != "admin" {
		return models.UserResponse{}, ErrNotAdmin
	}

	if err := s.userTerminalRepo.Delete(ctx, user.UUID, req.BusTerminalID); err != nil {
		return models.UserResponse{}, fmt.Errorf("failed to remove terminal association: %w", err)
	}

	remaining, err := s.userTerminalRepo.GetByUserID(ctx, user.UUID)
	if err != nil {
		return models.UserResponse{}, err
	}

	if len(remaining) == 0 {
		userRol, err := s.rolRepo.GetByName(ctx, "user")
		if err != nil {
			return models.UserResponse{}, fmt.Errorf("%w: %w", ErrRolNotFound, err)
		}
		user.RolID = userRol.UUID
		user.Rol = &userRol
		if err := s.userRepo.Update(ctx, &user); err != nil {
			return models.UserResponse{}, fmt.Errorf("failed to update user role: %w", err)
		}
	}

	return models.ToUserResponse(user), nil
}
