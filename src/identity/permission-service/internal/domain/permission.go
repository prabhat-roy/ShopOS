package domain

import (
	"errors"
	"time"
)

// Role represents a named collection of permissions that can be assigned to users.
type Role struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`        // e.g. "admin", "viewer"
	Description string    `json:"description"`
	Permissions []string  `json:"permissions"` // e.g. ["order:read","order:write"]
	CreatedAt   time.Time `json:"created_at"`
}

// UserRole represents the binding between a user and a role.
type UserRole struct {
	UserID     string    `json:"user_id"`
	RoleID     string    `json:"role_id"`
	RoleName   string    `json:"role_name"`
	AssignedAt time.Time `json:"assigned_at"`
}

// CheckRequest is the payload for a permission check.
type CheckRequest struct {
	UserID     string `json:"user_id"`
	Permission string `json:"permission"` // e.g. "order:read"
}

// CheckResponse is the result of a permission check.
type CheckResponse struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason"`
}

// Sentinel errors.
var ErrNotFound = errors.New("not found")
var ErrAlreadyAssigned = errors.New("role already assigned")
