package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	errorsService "tesina/backend/internal/errors"
	"tesina/backend/internal/models"
	"tesina/backend/internal/repository"
	"tesina/backend/internal/roles"
	"tesina/backend/internal/validators"
)

const (
	defaultCitiesLimit = 10
	maxCitiesLimit     = 100

	defaultPlatformsLimit = 10
	maxPlatformsLimit     = 100

	defaultUsersLimit = 10
	maxUsersLimit     = 100
)

type AdminService interface {
	// Cities
	ListCities(ctx context.Context, params models.ListCitiesParams) (models.ListCitiesResponse, error)
	CountCities(ctx context.Context) (int64, error)
	GetCity(ctx context.Context, postalCode string) (models.City, error)
	CreateCity(ctx context.Context, req models.CreateCityRequest) (models.City, error)
	UpdateCity(ctx context.Context, postalCode string, req models.UpdateCityRequest) (models.City, error)
	DeleteCity(ctx context.Context, postalCode string) error

	// Platforms (only admin's terminals)
	ListAllPlatforms(ctx context.Context, busTerminalID *uuid.UUID, params models.ListPlatformsParams) (models.ListPlatformsResponse, error)
	ListPlatforms(ctx context.Context, adminID uuid.UUID, busTerminalID *uuid.UUID, params models.ListPlatformsParams) (models.ListPlatformsResponse, error)
	CountAllPlatforms(ctx context.Context, busTerminalID *uuid.UUID) (int64, error)
	CountPlatforms(ctx context.Context, adminID uuid.UUID, busTerminalID *uuid.UUID) (int64, error)
	GetPlatformByCode(ctx context.Context, code int) (models.BusTerminalWithPlatformsResponse, error)
	GetPlatform(ctx context.Context, adminID uuid.UUID, code int) (models.BusTerminalWithPlatformsResponse, error)
	CreatePlatformDirect(ctx context.Context, req models.CreatePlatformRequest) (models.Platform, error)
	CreatePlatform(ctx context.Context, adminID uuid.UUID, req models.CreatePlatformRequest) (models.Platform, error)
	UpdatePlatformByCode(ctx context.Context, code int, req models.UpdatePlatformRequest) (models.Platform, error)
	UpdatePlatform(ctx context.Context, adminID uuid.UUID, code int, req models.UpdatePlatformRequest) (models.Platform, error)
	DeletePlatformByCode(ctx context.Context, code int) error
	DeletePlatform(ctx context.Context, adminID uuid.UUID, code int) error

	// User management
	ListUsers(ctx context.Context, params models.ListUsersParams) (models.ListUsersResponse, error)
	CountUsers(ctx context.Context) (int64, error)
	GetUserByEmail(ctx context.Context, email string) (models.AdminUserByEmailResponse, error)
	PromoteToAdminDirect(ctx context.Context, req models.PromoteAdminRequest) (models.UserResponse, error)
	PromoteToAdmin(ctx context.Context, adminID uuid.UUID, req models.PromoteAdminRequest) (models.UserResponse, error)
	DemoteAdminDirect(ctx context.Context, req models.DemoteAdminRequest) (models.UserResponse, error)
	DemoteAdmin(ctx context.Context, adminID uuid.UUID, req models.DemoteAdminRequest) (models.UserResponse, error)

	// Stats
	TotalTerminals(ctx context.Context) (int64, error)
	TotalPlatforms(ctx context.Context) (int64, error)
	TotalCities(ctx context.Context) (int64, error)
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

func (s *adminService) ListCities(ctx context.Context, params models.ListCitiesParams) (models.ListCitiesResponse, error) {
	page, limit, order := normalizePagination(params.Page, params.Limit, params.Order, defaultCitiesLimit, maxCitiesLimit)

	total, err := s.cityRepo.Count(ctx)
	if err != nil {
		return models.ListCitiesResponse{}, fmt.Errorf("failed to count cities: %w", err)
	}

	cities, err := s.cityRepo.ListPaginated(ctx, limit, (page-1)*limit, order)
	if err != nil {
		return models.ListCitiesResponse{}, fmt.Errorf("failed to list cities: %w", err)
	}
	if cities == nil {
		cities = []models.City{}
	}

	next, prev := pageLinks(total, page, limit)

	return models.ListCitiesResponse{
		Cities:        cities,
		Page:          page,
		Next:          next,
		Prev:          prev,
		Elements:      len(cities),
		TotalElements: total,
	}, nil
}

func (s *adminService) CountCities(ctx context.Context) (int64, error) {
	total, err := s.cityRepo.Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count cities: %w", err)
	}
	return total, nil
}

func (s *adminService) GetCity(ctx context.Context, postalCode string) (models.City, error) {
	city, err := s.cityRepo.GetByPostalCode(ctx, postalCode)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return models.City{}, errorsService.ErrCityNotFound
		}
		return models.City{}, err
	}
	return city, nil
}

func (s *adminService) CreateCity(ctx context.Context, req models.CreateCityRequest) (models.City, error) {
	if err := validators.ValidateCreateCityRequest(req); err != nil {
		return models.City{}, err
	}

	if _, err := s.cityRepo.GetByPostalCode(ctx, req.PostalCode); err == nil {
		return models.City{}, errorsService.ErrCityAlreadyExists
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
			return models.City{}, errorsService.ErrCityNotFound
		}
		return models.City{}, err
	}

	if req.Name != nil {
		city.Name = *req.Name
	}

	if req.PostalCode != nil && *req.PostalCode != postalCode {
		if _, err := s.cityRepo.GetByPostalCode(ctx, *req.PostalCode); err == nil {
			return models.City{}, errorsService.ErrCityAlreadyExists
		}

		if err := s.cityRepo.Update(ctx, &city); err != nil {
			return models.City{}, fmt.Errorf("failed to update city: %w", err)
		}

		if err := s.cityRepo.UpdatePostalCode(ctx, postalCode, *req.PostalCode); err != nil {
			return models.City{}, fmt.Errorf("failed to update postal code: %w", err)
		}
		city.PostalCode = *req.PostalCode
	} else {
		if err := s.cityRepo.Update(ctx, &city); err != nil {
			return models.City{}, fmt.Errorf("failed to update city: %w", err)
		}
	}

	return city, nil
}

func (s *adminService) DeleteCity(ctx context.Context, postalCode string) error {
	if _, err := s.cityRepo.GetByPostalCode(ctx, postalCode); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return errorsService.ErrCityNotFound
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
		return errorsService.ErrTerminalNotOwned
	}
	return nil
}

func (s *adminService) ListAllPlatforms(ctx context.Context, busTerminalID *uuid.UUID, params models.ListPlatformsParams) (models.ListPlatformsResponse, error) {
	page, limit, order := normalizePagination(params.Page, params.Limit, params.Order, defaultPlatformsLimit, maxPlatformsLimit)

	if busTerminalID != nil {
		bt, err := s.busTerminalRepo.GetByUUIDWithPlatforms(ctx, *busTerminalID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return models.ListPlatformsResponse{}, errorsService.ErrTerminalNotFound
			}
			return models.ListPlatformsResponse{}, err
		}
		return singleTerminalPlatformsPage(bt, page, limit), nil
	}

	total, err := s.busTerminalRepo.Count(ctx)
	if err != nil {
		return models.ListPlatformsResponse{}, fmt.Errorf("failed to count platforms: %w", err)
	}

	terminals, err := s.busTerminalRepo.ListWithPlatformsPaginated(ctx, limit, (page-1)*limit, order)
	if err != nil {
		return models.ListPlatformsResponse{}, fmt.Errorf("failed to list platforms: %w", err)
	}

	return platformsPage(terminals, total, page, limit), nil
}

func (s *adminService) CountAllPlatforms(ctx context.Context, busTerminalID *uuid.UUID) (int64, error) {
	if busTerminalID != nil {
		if _, err := s.busTerminalRepo.GetByUUID(ctx, *busTerminalID); err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return 0, errorsService.ErrTerminalNotFound
			}
			return 0, err
		}
		return 1, nil
	}

	total, err := s.busTerminalRepo.Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count platforms: %w", err)
	}
	return total, nil
}

func (s *adminService) ListPlatforms(ctx context.Context, adminID uuid.UUID, busTerminalID *uuid.UUID, params models.ListPlatformsParams) (models.ListPlatformsResponse, error) {
	page, limit, order := normalizePagination(params.Page, params.Limit, params.Order, defaultPlatformsLimit, maxPlatformsLimit)

	if busTerminalID != nil {
		if err := s.verifyTerminalOwnership(ctx, adminID, *busTerminalID); err != nil {
			return models.ListPlatformsResponse{}, err
		}

		bt, err := s.busTerminalRepo.GetByUUIDWithPlatforms(ctx, *busTerminalID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return models.ListPlatformsResponse{}, errorsService.ErrTerminalNotFound
			}
			return models.ListPlatformsResponse{}, err
		}
		return singleTerminalPlatformsPage(bt, page, limit), nil
	}

	ids, err := s.adminTerminalIDs(ctx, adminID)
	if err != nil {
		return models.ListPlatformsResponse{}, err
	}

	total, err := s.busTerminalRepo.CountByUUIDs(ctx, ids)
	if err != nil {
		return models.ListPlatformsResponse{}, fmt.Errorf("failed to count platforms: %w", err)
	}

	terminals, err := s.busTerminalRepo.ListByUUIDsPaginated(ctx, ids, limit, (page-1)*limit, order)
	if err != nil {
		return models.ListPlatformsResponse{}, fmt.Errorf("failed to list platforms: %w", err)
	}

	return platformsPage(terminals, total, page, limit), nil
}

func (s *adminService) CountPlatforms(ctx context.Context, adminID uuid.UUID, busTerminalID *uuid.UUID) (int64, error) {
	if busTerminalID != nil {
		if err := s.verifyTerminalOwnership(ctx, adminID, *busTerminalID); err != nil {
			return 0, err
		}
		return 1, nil
	}

	ids, err := s.adminTerminalIDs(ctx, adminID)
	if err != nil {
		return 0, err
	}

	total, err := s.busTerminalRepo.CountByUUIDs(ctx, ids)
	if err != nil {
		return 0, fmt.Errorf("failed to count platforms: %w", err)
	}
	return total, nil
}

// adminTerminalIDs devuelve los UUIDs de las terminales asignadas al admin.
func (s *adminService) adminTerminalIDs(ctx context.Context, adminID uuid.UUID) ([]uuid.UUID, error) {
	uts, err := s.userTerminalRepo.GetByUserID(ctx, adminID)
	if err != nil {
		return nil, err
	}

	ids := make([]uuid.UUID, len(uts))
	for i, ut := range uts {
		ids[i] = ut.BusTerminalID
	}
	return ids, nil
}

// platformsPage arma el payload paginado a partir de la página ya consultada.
func platformsPage(terminals []models.BusTerminal, total int64, page, limit int) models.ListPlatformsResponse {
	items := models.ToBusTerminalWithPlatformsResponse(terminals)
	if items == nil {
		items = []models.BusTerminalWithPlatformsResponse{}
	}
	next, prev := pageLinks(total, page, limit)

	return models.ListPlatformsResponse{
		Platforms:     items,
		Page:          page,
		Next:          next,
		Prev:          prev,
		Elements:      len(items),
		TotalElements: total,
	}
}

// singleTerminalPlatformsPage arma el payload cuando se filtra por bus_terminal_id,
// caso en el que el resultado es siempre una única terminal.
func singleTerminalPlatformsPage(bt models.BusTerminal, page, limit int) models.ListPlatformsResponse {
	if page > 1 {
		return platformsPage(nil, 1, page, limit)
	}
	return platformsPage([]models.BusTerminal{bt}, 1, page, limit)
}

func (s *adminService) GetPlatformByCode(ctx context.Context, code int) (models.BusTerminalWithPlatformsResponse, error) {
	platform, err := s.platformRepo.GetByCode(ctx, code)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return models.BusTerminalWithPlatformsResponse{}, errorsService.ErrPlatformNotFound
		}
		return models.BusTerminalWithPlatformsResponse{}, err
	}

	bt, err := s.busTerminalRepo.GetByUUID(ctx, platform.BusTerminalID)
	if err != nil {
		return models.BusTerminalWithPlatformsResponse{}, err
	}
	bt.Platforms = []models.Platform{platform}

	return models.ToBusTerminalWithPlatformsResponse([]models.BusTerminal{bt})[0], nil
}

func (s *adminService) GetPlatform(ctx context.Context, adminID uuid.UUID, code int) (models.BusTerminalWithPlatformsResponse, error) {
	platform, err := s.platformRepo.GetByCode(ctx, code)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return models.BusTerminalWithPlatformsResponse{}, errorsService.ErrPlatformNotFound
		}
		return models.BusTerminalWithPlatformsResponse{}, err
	}

	if err := s.verifyTerminalOwnership(ctx, adminID, platform.BusTerminalID); err != nil {
		return models.BusTerminalWithPlatformsResponse{}, err
	}

	bt, err := s.busTerminalRepo.GetByUUID(ctx, platform.BusTerminalID)
	if err != nil {
		return models.BusTerminalWithPlatformsResponse{}, err
	}
	bt.Platforms = []models.Platform{platform}

	return models.ToBusTerminalWithPlatformsResponse([]models.BusTerminal{bt})[0], nil
}

func (s *adminService) CreatePlatformDirect(ctx context.Context, req models.CreatePlatformRequest) (models.Platform, error) {
	if err := validators.ValidateCreatePlatformRequest(req); err != nil {
		return models.Platform{}, err
	}

	if _, err := s.busTerminalRepo.GetByUUID(ctx, req.BusTerminalID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return models.Platform{}, errorsService.ErrTerminalNotFound
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
	if err := validators.ValidateCreatePlatformRequest(req); err != nil {
		return models.Platform{}, err
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
			return models.Platform{}, errorsService.ErrPlatformNotFound
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
			return models.Platform{}, errorsService.ErrPlatformNotFound
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
			return errorsService.ErrPlatformNotFound
		}
		return err
	}
	return s.platformRepo.Delete(ctx, code)
}

func (s *adminService) DeletePlatform(ctx context.Context, adminID uuid.UUID, code int) error {
	platform, err := s.platformRepo.GetByCode(ctx, code)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return errorsService.ErrPlatformNotFound
		}
		return err
	}

	if err := s.verifyTerminalOwnership(ctx, adminID, platform.BusTerminalID); err != nil {
		return err
	}

	return s.platformRepo.Delete(ctx, code)
}

// --- User management ---

func (s *adminService) ListUsers(ctx context.Context, params models.ListUsersParams) (models.ListUsersResponse, error) {
	page, limit, order := normalizePagination(params.Page, params.Limit, params.Order, defaultUsersLimit, maxUsersLimit)

	total, err := s.userRepo.Count(ctx, params.Search)
	if err != nil {
		return models.ListUsersResponse{}, fmt.Errorf("failed to count users: %w", err)
	}

	users, err := s.userRepo.ListPaginated(ctx, params.Search, limit, (page-1)*limit, order)
	if err != nil {
		return models.ListUsersResponse{}, fmt.Errorf("failed to list users: %w", err)
	}

	ids := make([]uuid.UUID, len(users))
	for i := range users {
		ids[i] = users[i].UUID
	}
	terminalsByUser, err := s.userTerminalRepo.ListTerminalRefsByUserIDs(ctx, ids)
	if err != nil {
		return models.ListUsersResponse{}, fmt.Errorf("failed to list user terminals: %w", err)
	}

	items := make([]models.UserListItem, len(users))
	for i := range users {
		items[i] = models.ToUserListItem(users[i], terminalsByUser[users[i].UUID])
	}

	next, prev := pageLinks(total, page, limit)

	return models.ListUsersResponse{
		Users:         items,
		Page:          page,
		Next:          next,
		Prev:          prev,
		Elements:      len(items),
		TotalElements: total,
	}, nil
}

func (s *adminService) CountUsers(ctx context.Context) (int64, error) {
	total, err := s.userRepo.Count(ctx, "")
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}
	return total, nil
}

func (s *adminService) GetUserByEmail(ctx context.Context, email string) (models.AdminUserByEmailResponse, error) {
	var err error
	email, err = validators.ValidateAdminEmail(email)
	if err != nil {
		return models.AdminUserByEmailResponse{}, err
	}

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return models.AdminUserByEmailResponse{}, errorsService.ErrUserNotFound
		}
		return models.AdminUserByEmailResponse{}, err
	}

	var terminals []string
	if user.Rol != nil && user.Rol.Name == roles.Admin {
		userTerminals, err := s.userTerminalRepo.GetByUserID(ctx, user.UUID)
		if err != nil {
			return models.AdminUserByEmailResponse{}, err
		}
		if len(userTerminals) > 0 {
			ids := make([]uuid.UUID, len(userTerminals))
			for i := range userTerminals {
				ids[i] = userTerminals[i].BusTerminalID
			}
			busTerminals, err := s.busTerminalRepo.ListByUUIDs(ctx, ids)
			if err != nil {
				return models.AdminUserByEmailResponse{}, err
			}
			terminals = make([]string, len(busTerminals))
			for i := range busTerminals {
				terminals[i] = busTerminals[i].Name
			}
		}
	}

	return models.ToAdminUserByEmailResponse(user, terminals), nil
}

func (s *adminService) PromoteToAdminDirect(ctx context.Context, req models.PromoteAdminRequest) (models.UserResponse, error) {
	if err := validators.ValidatePromoteAdminRequest(req); err != nil {
		return models.UserResponse{}, err
	}

	if _, err := s.busTerminalRepo.GetByUUID(ctx, req.BusTerminalID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return models.UserResponse{}, errorsService.ErrTerminalNotFound
		}
		return models.UserResponse{}, err
	}

	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return models.UserResponse{}, errorsService.ErrUserNotFound
		}
		return models.UserResponse{}, err
	}

	if user.Rol != nil && user.Rol.Name == roles.SuperAdmin {
		return models.UserResponse{}, errorsService.ErrAlreadySuperAdmin
	}

	adminRol, err := s.rolRepo.GetByName(ctx, roles.Admin)
	if err != nil {
		return models.UserResponse{}, fmt.Errorf("%w: %w", errorsService.ErrRolNotFound, err)
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
	if err := validators.ValidatePromoteAdminRequest(req); err != nil {
		return models.UserResponse{}, err
	}

	if err := s.verifyTerminalOwnership(ctx, adminID, req.BusTerminalID); err != nil {
		return models.UserResponse{}, err
	}

	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return models.UserResponse{}, errorsService.ErrUserNotFound
		}
		return models.UserResponse{}, err
	}

	adminRol, err := s.rolRepo.GetByName(ctx, roles.Admin)
	if err != nil {
		return models.UserResponse{}, fmt.Errorf("%w: %w", errorsService.ErrRolNotFound, err)
	}

	if user.Rol != nil && user.Rol.Name == roles.SuperAdmin {
		return models.UserResponse{}, errorsService.ErrAlreadySuperAdmin
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
	if err := validators.ValidateDemoteAdminRequest(req); err != nil {
		return models.UserResponse{}, err
	}

	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return models.UserResponse{}, errorsService.ErrUserNotFound
		}
		return models.UserResponse{}, err
	}

	if user.Rol == nil || user.Rol.Name != roles.Admin {
		return models.UserResponse{}, errorsService.ErrNotAdmin
	}

	if err := s.userTerminalRepo.Delete(ctx, user.UUID, req.BusTerminalID); err != nil {
		return models.UserResponse{}, fmt.Errorf("failed to remove terminal association: %w", err)
	}

	remaining, err := s.userTerminalRepo.GetByUserID(ctx, user.UUID)
	if err != nil {
		return models.UserResponse{}, err
	}

	if len(remaining) == 0 {
		userRol, err := s.rolRepo.GetByName(ctx, roles.User)
		if err != nil {
			return models.UserResponse{}, fmt.Errorf("%w: %w", errorsService.ErrRolNotFound, err)
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
	if err := validators.ValidateDemoteAdminRequest(req); err != nil {
		return models.UserResponse{}, err
	}

	if err := s.verifyTerminalOwnership(ctx, adminID, req.BusTerminalID); err != nil {
		return models.UserResponse{}, err
	}

	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return models.UserResponse{}, errorsService.ErrUserNotFound
		}
		return models.UserResponse{}, err
	}

	if user.UUID == adminID {
		return models.UserResponse{}, errorsService.ErrCannotDemoteSelf
	}

	if user.Rol == nil || user.Rol.Name != roles.Admin {
		return models.UserResponse{}, errorsService.ErrNotAdmin
	}

	if err := s.userTerminalRepo.Delete(ctx, user.UUID, req.BusTerminalID); err != nil {
		return models.UserResponse{}, fmt.Errorf("failed to remove terminal association: %w", err)
	}

	remaining, err := s.userTerminalRepo.GetByUserID(ctx, user.UUID)
	if err != nil {
		return models.UserResponse{}, err
	}

	if len(remaining) == 0 {
		userRol, err := s.rolRepo.GetByName(ctx, roles.User)
		if err != nil {
			return models.UserResponse{}, fmt.Errorf("%w: %w", errorsService.ErrRolNotFound, err)
		}
		user.RolID = userRol.UUID
		user.Rol = &userRol
		if err := s.userRepo.Update(ctx, &user); err != nil {
			return models.UserResponse{}, fmt.Errorf("failed to update user role: %w", err)
		}
	}

	return models.ToUserResponse(user), nil
}

// --- Stats ---

// TotalTerminals devuelve la cantidad de terminales registradas.
func (s *adminService) TotalTerminals(ctx context.Context) (int64, error) {
	total, err := s.busTerminalRepo.Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count terminals: %w", err)
	}
	return total, nil
}

// TotalPlatforms devuelve la cantidad de plataformas (andenes) registradas.
func (s *adminService) TotalPlatforms(ctx context.Context) (int64, error) {
	total, err := s.platformRepo.Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count platforms: %w", err)
	}
	return total, nil
}

// TotalCities devuelve la cantidad de ciudades registradas.
func (s *adminService) TotalCities(ctx context.Context) (int64, error) {
	total, err := s.cityRepo.Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count cities: %w", err)
	}
	return total, nil
}
