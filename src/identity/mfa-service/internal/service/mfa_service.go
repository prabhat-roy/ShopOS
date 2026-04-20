package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/pquerna/otp/totp"
	"github.com/shopos/mfa-service/internal/domain"
)

// Storer is the persistence interface consumed by MFAService.
type Storer interface {
	Save(ctx context.Context, setup *domain.MFASetup) error
	Get(ctx context.Context, userID string) (*domain.MFASetup, error)
	UpdateStatus(ctx context.Context, userID string, status domain.MFAStatus) error
	UseBackupCode(ctx context.Context, userID, code string) error
}

// MFAService implements all business logic for multi-factor authentication.
type MFAService struct {
	store   Storer
	issuer  string
}

// New returns a ready MFAService.
func New(store Storer, issuer string) *MFAService {
	if issuer == "" {
		issuer = "ShopOS"
	}
	return &MFAService{store: store, issuer: issuer}
}

// Enroll generates a new TOTP secret and 8 random backup codes for the user,
// persists the setup in "pending" status, and returns the setup (including the
// plaintext backup codes for one-time display).
//
// Returns domain.ErrAlreadyEnabled if MFA is already active for the user.
func (s *MFAService) Enroll(ctx context.Context, userID string) (*domain.MFASetup, error) {
	existing, err := s.store.Get(ctx, userID)
	if err != nil && err != domain.ErrNotFound {
		return nil, fmt.Errorf("service: enroll get: %w", err)
	}
	if existing != nil && existing.Status == domain.MFAEnabled {
		return nil, domain.ErrAlreadyEnabled
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      s.issuer,
		AccountName: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("service: generate totp key: %w", err)
	}

	backupCodes, err := generateBackupCodes(8)
	if err != nil {
		return nil, fmt.Errorf("service: generate backup codes: %w", err)
	}

	setup := &domain.MFASetup{
		UserID:      userID,
		Secret:      key.Secret(),
		QRCodeURL:   key.URL(),
		BackupCodes: backupCodes,
		Status:      domain.MFAPending,
	}

	if err := s.store.Save(ctx, setup); err != nil {
		return nil, fmt.Errorf("service: save setup: %w", err)
	}

	return setup, nil
}

// Confirm verifies the supplied TOTP code against the user's pending secret and
// transitions the status to "enabled".
//
// Returns domain.ErrNotFound if MFA has not been enrolled, and
// domain.ErrInvalidCode if the code does not validate.
func (s *MFAService) Confirm(ctx context.Context, userID, code string) error {
	setup, err := s.store.Get(ctx, userID)
	if err != nil {
		return err // propagates ErrNotFound
	}

	if !totp.Validate(code, setup.Secret) {
		return domain.ErrInvalidCode
	}

	if err := s.store.UpdateStatus(ctx, userID, domain.MFAEnabled); err != nil {
		return fmt.Errorf("service: update status: %w", err)
	}
	return nil
}

// Verify validates a login-time code. It accepts either a current TOTP code or
// one of the user's unused backup codes.
//
// Returns domain.ErrNotFound if MFA is not configured, and
// domain.ErrInvalidCode if neither form of code matches.
func (s *MFAService) Verify(ctx context.Context, req *domain.VerifyRequest) (*domain.VerifyResponse, error) {
	setup, err := s.store.Get(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	if setup.Status != domain.MFAEnabled {
		return nil, domain.ErrNotFound
	}

	// Try TOTP first.
	if totp.Validate(req.Code, setup.Secret) {
		return &domain.VerifyResponse{Valid: true, IsBackupCode: false}, nil
	}

	// Attempt to consume a backup code.
	err = s.store.UseBackupCode(ctx, req.UserID, req.Code)
	if err == nil {
		return &domain.VerifyResponse{Valid: true, IsBackupCode: true}, nil
	}
	if err == domain.ErrNotFound {
		return nil, domain.ErrInvalidCode
	}
	return nil, fmt.Errorf("service: use backup code: %w", err)
}

// Disable sets the user's MFA status to "disabled".
// Returns domain.ErrNotFound if MFA was never configured.
func (s *MFAService) Disable(ctx context.Context, userID string) error {
	if err := s.store.UpdateStatus(ctx, userID, domain.MFADisabled); err != nil {
		return err
	}
	return nil
}

// GetStatus returns the current MFAStatus for the given user.
// Returns domain.ErrNotFound if MFA has never been configured.
func (s *MFAService) GetStatus(ctx context.Context, userID string) (domain.MFAStatus, error) {
	setup, err := s.store.Get(ctx, userID)
	if err != nil {
		return "", err
	}
	return setup.Status, nil
}

// generateBackupCodes returns n random 8-character hex strings.
func generateBackupCodes(n int) ([]string, error) {
	codes := make([]string, n)
	buf := make([]byte, 4) // 4 bytes → 8 hex chars
	for i := range codes {
		if _, err := rand.Read(buf); err != nil {
			return nil, err
		}
		codes[i] = hex.EncodeToString(buf)
	}
	return codes, nil
}
