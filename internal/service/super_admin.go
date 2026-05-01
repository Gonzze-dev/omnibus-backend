package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"tesina/backend/internal/models"
	"tesina/backend/internal/repository"
	"tesina/backend/internal/validators"
	errorsService "tesina/backend/internal/errors"
)

type SuperAdminService interface {
	// Terminals
	ListTerminals(ctx context.Context) ([]models.BusTerminal, error)
	GetTerminal(ctx context.Context, id uuid.UUID) (models.BusTerminal, error)
	CreateTerminal(ctx context.Context, req models.CreateBusTerminalRequest) (models.BusTerminal, error)
	UpdateTerminal(ctx context.Context, id uuid.UUID, req models.UpdateBusTerminalRequest) (models.BusTerminal, error)
	DeleteTerminal(ctx context.Context, id uuid.UUID) error

	// User management (super-only)
	PromoteToSuper(ctx context.Context, req models.PromoteSuperRequest) (models.UserResponse, error)
	DemoteSuper(ctx context.Context, req models.DemoteSuperRequest) (models.UserResponse, error)
}

type superAdminService struct {
	cityRepo              repository.CityRepository
	busTerminalRepo       repository.BusTerminalRepository
	userRepo              repository.UserRepository
	rolRepo               repository.RolRepository
	userTerminalRepo      repository.UserTerminalRepository
	externalTerminalCheck BusTicketService
}

func NewSuperAdminService(
	cityRepo repository.CityRepository,
	busTerminalRepo repository.BusTerminalRepository,
	userRepo repository.UserRepository,
	rolRepo repository.RolRepository,
	userTerminalRepo repository.UserTerminalRepository,
	externalTerminalCheck BusTicketService,
) *superAdminService {
	return &superAdminService{
		cityRepo:              cityRepo,
		busTerminalRepo:       busTerminalRepo,
		userRepo:              userRepo,
		rolRepo:               rolRepo,
		userTerminalRepo:      userTerminalRepo,
		externalTerminalCheck: externalTerminalCheck,
	}
}

// --- Terminals ---

func (s *superAdminService) ListTerminals(ctx context.Context) ([]models.BusTerminal, error) {
	return s.busTerminalRepo.List(ctx)
}

func (s *superAdminService) GetTerminal(ctx context.Context, id uuid.UUID) (models.BusTerminal, error) {
	terminal, err := s.busTerminalRepo.GetByUUID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return models.BusTerminal{}, errorsService.ErrTerminalNotFound
		}
		return models.BusTerminal{}, err
	}
	return terminal, nil
}

func (s *superAdminService) CreateTerminal(ctx context.Context, req models.CreateBusTerminalRequest) (models.BusTerminal, error) {
	if err := validators.ValidateCreateBusTerminalRequest(req); err != nil {
		return models.BusTerminal{}, err
	}

	if _, err := s.busTerminalRepo.GetByExternalTerminalID(ctx, req.ExternalTerminalID); err == nil {
		return models.BusTerminal{}, errorsService.ErrExternalTerminalIDAlreadyUsed
	} else if !errors.Is(err, repository.ErrNotFound) {
		return models.BusTerminal{}, err
	}

	exists, err := s.externalTerminalCheck.ExternalTerminalExists(ctx, req.ExternalTerminalID)
	if err != nil {
		return models.BusTerminal{}, err
	}
	if !exists {
		return models.BusTerminal{}, errorsService.ErrInvalidExternalTerminalID
	}

	if _, err := s.cityRepo.GetByPostalCode(ctx, req.PostalCode); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return models.BusTerminal{}, errorsService.ErrCityNotFound
		}
		return models.BusTerminal{}, err
	}

	extID := req.ExternalTerminalID
	terminal := models.BusTerminal{
		UUID:               uuid.New(),
		ExternalTerminalID: &extID,
		PostalCode:         req.PostalCode,
		Name:               req.Name,
	}
	if err := s.busTerminalRepo.Create(ctx, &terminal); err != nil {
		return models.BusTerminal{}, fmt.Errorf("failed to create terminal: %w", err)
	}
	return terminal, nil
}

func (s *superAdminService) UpdateTerminal(ctx context.Context, id uuid.UUID, req models.UpdateBusTerminalRequest) (models.BusTerminal, error) {
	if err := validators.ValidateUpdateBusTerminalRequest(req); err != nil {
		return models.BusTerminal{}, err
	}

	terminal, err := s.busTerminalRepo.GetByUUID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return models.BusTerminal{}, errorsService.ErrTerminalNotFound
		}
		return models.BusTerminal{}, err
	}

	if req.PostalCode != nil {
		if _, err := s.cityRepo.GetByPostalCode(ctx, *req.PostalCode); err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return models.BusTerminal{}, errorsService.ErrCityNotFound
			}
			return models.BusTerminal{}, err
		}
		terminal.PostalCode = *req.PostalCode
	}
	if req.Name != nil {
		terminal.Name = *req.Name
	}

	if req.ExternalTerminalID != nil {
		newExt := *req.ExternalTerminalID
		sameAsCurrent := terminal.ExternalTerminalID != nil && *terminal.ExternalTerminalID == newExt
		if !sameAsCurrent {
			if existing, err := s.busTerminalRepo.GetByExternalTerminalID(ctx, newExt); err == nil {
				if existing.UUID != terminal.UUID {
					return models.BusTerminal{}, errorsService.ErrExternalTerminalIDAlreadyUsed
				}
			} else if !errors.Is(err, repository.ErrNotFound) {
				return models.BusTerminal{}, err
			}

			exists, err := s.externalTerminalCheck.ExternalTerminalExists(ctx, newExt)
			if err != nil {
				return models.BusTerminal{}, err
			}
			if !exists {
				return models.BusTerminal{}, errorsService.ErrInvalidExternalTerminalID
			}
		}

		extCopy := newExt
		terminal.ExternalTerminalID = &extCopy
	}

	if err := s.busTerminalRepo.Update(ctx, &terminal); err != nil {
		return models.BusTerminal{}, fmt.Errorf("failed to update terminal: %w", err)
	}
	return terminal, nil
}

func (s *superAdminService) DeleteTerminal(ctx context.Context, id uuid.UUID) error {
	if _, err := s.busTerminalRepo.GetByUUID(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return errorsService.ErrTerminalNotFound
		}
		return err
	}
	return s.busTerminalRepo.Delete(ctx, id)
}

// --- User management (super-only) ---

func (s *superAdminService) PromoteToSuper(ctx context.Context, req models.PromoteSuperRequest) (models.UserResponse, error) {
	if err := validators.ValidatePromoteSuperRequest(req); err != nil {
		return models.UserResponse{}, err
	}

	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return models.UserResponse{}, errorsService.ErrUserNotFound
		}
		return models.UserResponse{}, err
	}

	if user.Rol != nil && user.Rol.Name == "super_admin" {
		return models.UserResponse{}, errorsService.ErrAlreadySuperAdmin
	}

	superRol, err := s.rolRepo.GetByName(ctx, "super_admin")
	if err != nil {
		return models.UserResponse{}, fmt.Errorf("%w: %w", errorsService.ErrRolNotFound, err)
	}

	user.RolID = superRol.UUID
	user.Rol = &superRol
	if err := s.userRepo.Update(ctx, &user); err != nil {
		return models.UserResponse{}, fmt.Errorf("failed to update user role: %w", err)
	}

	uts, err := s.userTerminalRepo.GetByUserID(ctx, user.UUID)
	if err == nil {
		for _, ut := range uts {
			_ = s.userTerminalRepo.Delete(ctx, ut.UserID, ut.BusTerminalID)
		}
	}

	return models.ToUserResponse(user), nil
}

func (s *superAdminService) DemoteSuper(ctx context.Context, req models.DemoteSuperRequest) (models.UserResponse, error) {
	if err := validators.ValidateDemoteSuperRequest(req); err != nil {
		return models.UserResponse{}, err
	}

	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return models.UserResponse{}, errorsService.ErrUserNotFound
		}
		return models.UserResponse{}, err
	}

	if user.Rol == nil || user.Rol.Name != "super_admin" {
		return models.UserResponse{}, errorsService.ErrNotSuperAdmin
	}

	userRol, err := s.rolRepo.GetByName(ctx, "user")
	if err != nil {
		return models.UserResponse{}, fmt.Errorf("%w: %w", errorsService.ErrRolNotFound, err)
	}

	user.RolID = userRol.UUID
	user.Rol = &userRol
	if err := s.userRepo.Update(ctx, &user); err != nil {
		return models.UserResponse{}, fmt.Errorf("failed to update user role: %w", err)
	}

	return models.ToUserResponse(user), nil
}
