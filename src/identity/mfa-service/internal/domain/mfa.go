package domain

import (
	"errors"
	"time"
)

// MFAStatus represents the current state of MFA for a user.
type MFAStatus string

const (
	MFAEnabled  MFAStatus = "enabled"
	MFADisabled MFAStatus = "disabled"
	MFAPending  MFAStatus = "pending"
)

// MFASetup holds all MFA configuration for a single user.
type MFASetup struct {
	UserID      string    // Owning user
	Secret      string    // Base32-encoded TOTP secret
	QRCodeURL   string    // otpauth:// URI suitable for QR-code generation
	BackupCodes []string  // Plaintext backup codes (only populated on initial enroll)
	Status      MFAStatus // pending → enabled → disabled lifecycle
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// VerifyRequest is the input for a verify call.
type VerifyRequest struct {
	UserID string
	Code   string
}

// VerifyResponse is the result of a successful verification.
type VerifyResponse struct {
	Valid        bool
	IsBackupCode bool
}

// Sentinel errors returned by the service layer.
var (
	ErrNotFound      = errors.New("mfa not configured")
	ErrAlreadyEnabled = errors.New("mfa already enabled")
	ErrInvalidCode   = errors.New("invalid or expired code")
)
