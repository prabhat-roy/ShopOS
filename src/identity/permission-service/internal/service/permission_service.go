package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopos/permission-service/internal/domain"
)

// Storer is the data-access interface the service layer depends on.
type Storer interface {
	CreateRole(ctx context.Context, role *domain.Role) (*domain.Role, error)
	GetRole(ctx context.Context, id string) (*domain.Role, error)
	ListRoles(ctx context.Context) ([]*domain.Role, error)
	DeleteRole(ctx context.Context, id string) error

	AssignRole(ctx context.Context, userID, roleID string) error
	RevokeRole(ctx context.Context, userID, roleID string) error
	GetUserRoles(ctx context.Context, userID string) ([]*domain.UserRole, error)
	GetUserPermissions(ctx context.Context, userID string) ([]string, error)
}

// Service implements the business logic for the permission service.
type Service struct {
	store Storer
}

// New constructs a Service with the provided Storer implementation.
func New(store Storer) *Service {
	return &Service{store: store}
}

// CreateRole validates input and persists a new role.
func (s *Service) CreateRole(ctx context.Context, name, description string, permissions []string) (*domain.Role, error) {
	if name == "" {
		return nil, fmt.Errorf("role name is required")
	}
	if permissions == nil {
		permissions = []string{}
	}
	role := &domain.Role{
		ID:          uuid.NewString(),
		Name:        name,
		Description: description,
		Permissions: permissions,
		CreatedAt:   time.Now().UTC(),
	}
	return s.store.CreateRole(ctx, role)
}

// GetRole retrieves a role by ID.
func (s *Service) GetRole(ctx context.Context, id string) (*domain.Role, error) {
	if id == "" {
		return nil, fmt.Errorf("role id is required")
	}
	return s.store.GetRole(ctx, id)
}

// ListRoles returns all known roles.
func (s *Service) ListRoles(ctx context.Context) ([]*domain.Role, error) {
	return s.store.ListRoles(ctx)
}

// DeleteRole removes a role by ID.
func (s *Service) DeleteRole(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("role id is required")
	}
	return s.store.DeleteRole(ctx, id)
}

// AssignRole binds a role to a user.
func (s *Service) AssignRole(ctx context.Context, userID, roleID string) error {
	if userID == "" {
		return fmt.Errorf("user_id is required")
	}
	if roleID == "" {
		return fmt.Errorf("role_id is required")
	}
	return s.store.AssignRole(ctx, userID, roleID)
}

// RevokeRole removes a role binding from a user.
func (s *Service) RevokeRole(ctx context.Context, userID, roleID string) error {
	if userID == "" {
		return fmt.Errorf("user_id is required")
	}
	if roleID == "" {
		return fmt.Errorf("role_id is required")
	}
	return s.store.RevokeRole(ctx, userID, roleID)
}

// GetUserRoles returns all roles assigned to userID.
func (s *Service) GetUserRoles(ctx context.Context, userID string) ([]*domain.UserRole, error) {
	if userID == "" {
		return nil, fmt.Errorf("user_id is required")
	}
	return s.store.GetUserRoles(ctx, userID)
}

// Check evaluates whether a user holds the requested permission.
func (s *Service) Check(ctx context.Context, req domain.CheckRequest) domain.CheckResponse {
	if req.UserID == "" || req.Permission == "" {
		return domain.CheckResponse{Allowed: false, Reason: "user_id and permission are required"}
	}

	perms, err := s.store.GetUserPermissions(ctx, req.UserID)
	if err != nil {
		return domain.CheckResponse{Allowed: false, Reason: "internal error evaluating permissions"}
	}

	for _, p := range perms {
		if p == req.Permission {
			return domain.CheckResponse{Allowed: true, Reason: "permission granted"}
		}
	}
	return domain.CheckResponse{Allowed: false, Reason: fmt.Sprintf("user does not have permission %q", req.Permission)}
}
