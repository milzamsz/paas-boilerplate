package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"paas-core/apps/api/internal/auth"
	apiErrors "paas-core/apps/api/internal/errors"
	"paas-core/apps/api/internal/model"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailExists        = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// FilterParams for listing users.
type FilterParams struct {
	Search string
	Sort   string
	Order  string
}

// UpdateUserRequest is the DTO for user updates.
type UpdateUserRequest struct {
	Name      string `json:"name" binding:"omitempty,min=2,max=100"`
	Email     string `json:"email" binding:"omitempty,email"`
	AvatarURL string `json:"avatar_url" binding:"omitempty,url"`
}

// Service defines the user service interface.
type Service interface {
	RegisterUser(ctx *gin.Context, req auth.RegisterRequest) (*auth.UserResponse, []string, error)
	AuthenticateUser(ctx *gin.Context, req auth.LoginRequest) (*auth.UserResponse, []string, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*auth.UserResponse, error)
	UpdateUser(ctx context.Context, id uuid.UUID, req UpdateUserRequest) (*auth.UserResponse, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
	ListUsers(ctx context.Context, filters FilterParams, page, perPage int) ([]auth.UserResponse, int64, error)
}

type service struct {
	repo Repository
}

// NewService creates a new user service.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) RegisterUser(ctx *gin.Context, req auth.RegisterRequest) (*auth.UserResponse, []string, error) {
	// NIST password validation (min 12 chars, complexity, blocklist)
	if err := ValidatePasswordNIST(req.Password); err != nil {
		return nil, nil, apiErrors.BadRequest(err.Error())
	}

	existing, err := s.repo.FindByEmail(ctx.Request.Context(), req.Email)
	if err != nil {
		return nil, nil, apiErrors.InternalServerError(fmt.Errorf("failed to check email: %w", err))
	}
	if existing != nil {
		return nil, nil, apiErrors.Conflict("Email already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, apiErrors.InternalServerError(fmt.Errorf("failed to hash password: %w", err))
	}

	user := &model.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
	}

	err = s.repo.Transaction(ctx.Request.Context(), func(txCtx context.Context) error {
		if err := s.repo.Create(txCtx, user); err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}
		if err := s.repo.AssignRole(txCtx, user.ID, model.RoleUser); err != nil {
			return fmt.Errorf("failed to assign role: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, nil, apiErrors.InternalServerError(err)
	}

	// Reload with roles
	user, err = s.repo.FindByID(ctx.Request.Context(), user.ID)
	if err != nil || user == nil {
		return nil, nil, apiErrors.InternalServerError(fmt.Errorf("failed to reload user"))
	}

	roles := extractRoleNames(user.Roles)
	return toUserResponse(user), roles, nil
}

func (s *service) AuthenticateUser(ctx *gin.Context, req auth.LoginRequest) (*auth.UserResponse, []string, error) {
	user, err := s.repo.FindByEmail(ctx.Request.Context(), req.Email)
	if err != nil {
		return nil, nil, apiErrors.InternalServerError(fmt.Errorf("failed to find user: %w", err))
	}
	if user == nil {
		return nil, nil, apiErrors.Unauthorized("Invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, nil, apiErrors.Unauthorized("Invalid email or password")
	}

	roles := extractRoleNames(user.Roles)
	return toUserResponse(user), roles, nil
}

func (s *service) GetUserByID(ctx context.Context, id uuid.UUID) (*auth.UserResponse, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apiErrors.InternalServerError(err)
	}
	if user == nil {
		return nil, apiErrors.NotFound("User not found")
	}
	return toUserResponse(user), nil
}

func (s *service) UpdateUser(ctx context.Context, id uuid.UUID, req UpdateUserRequest) (*auth.UserResponse, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apiErrors.InternalServerError(err)
	}
	if user == nil {
		return nil, apiErrors.NotFound("User not found")
	}

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Email != "" {
		existing, err := s.repo.FindByEmail(ctx, req.Email)
		if err != nil {
			return nil, apiErrors.InternalServerError(err)
		}
		if existing != nil && existing.ID != user.ID {
			return nil, apiErrors.Conflict("Email already exists")
		}
		user.Email = req.Email
	}
	if req.AvatarURL != "" {
		user.AvatarURL = req.AvatarURL
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, apiErrors.InternalServerError(err)
	}
	return toUserResponse(user), nil
}

func (s *service) DeleteUser(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return apiErrors.InternalServerError(err)
	}
	return nil
}

func (s *service) ListUsers(ctx context.Context, filters FilterParams, page, perPage int) ([]auth.UserResponse, int64, error) {
	users, total, err := s.repo.ListAllUsers(ctx, filters, page, perPage)
	if err != nil {
		return nil, 0, apiErrors.InternalServerError(err)
	}
	var responses []auth.UserResponse
	for _, u := range users {
		responses = append(responses, *toUserResponse(&u))
	}
	return responses, total, nil
}

// --- Helpers ---

func toUserResponse(u *model.User) *auth.UserResponse {
	return &auth.UserResponse{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		AvatarURL: u.AvatarURL,
		Roles:     extractRoleNames(u.Roles),
		CreatedAt: u.CreatedAt,
	}
}

func extractRoleNames(roles []model.Role) []string {
	names := make([]string, 0, len(roles))
	for _, r := range roles {
		names = append(names, r.Name)
	}
	return names
}
