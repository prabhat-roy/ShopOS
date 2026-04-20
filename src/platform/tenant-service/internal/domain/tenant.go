package domain

import (
	"errors"
	"time"
)

// TenantStatus represents the lifecycle state of a tenant.
type TenantStatus string

const (
	StatusActive    TenantStatus = "active"
	StatusSuspended TenantStatus = "suspended"
	StatusDeleted   TenantStatus = "deleted"
)

// Plan represents the subscription tier of a tenant.
type Plan string

const (
	PlanStarter    Plan = "starter"
	PlanPro        Plan = "pro"
	PlanEnterprise Plan = "enterprise"
)

// Tenant is the core aggregate representing a platform tenant.
type Tenant struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Slug       string            `json:"slug"`
	OwnerEmail string            `json:"owner_email"`
	Plan       Plan              `json:"plan"`
	Status     TenantStatus      `json:"status"`
	Settings   map[string]string `json:"settings"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
}

// CreateTenantRequest carries the fields required to create a new tenant.
type CreateTenantRequest struct {
	Name       string `json:"name"`
	Slug       string `json:"slug"`
	OwnerEmail string `json:"owner_email"`
	Plan       Plan   `json:"plan"`
}

// UpdateTenantRequest carries optional fields for updating a tenant.
// Nil pointer fields are left unchanged.
type UpdateTenantRequest struct {
	Name     *string           `json:"name,omitempty"`
	Plan     *Plan             `json:"plan,omitempty"`
	Status   *TenantStatus     `json:"status,omitempty"`
	Settings map[string]string `json:"settings,omitempty"`
}

// Sentinel errors returned by the store and service layers.
var (
	ErrNotFound  = errors.New("not found")
	ErrSlugTaken = errors.New("slug already taken")
)
