package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopos/device-fingerprint-service/internal/domain"
	"github.com/shopos/device-fingerprint-service/internal/store"
)

// trustScoreNew is assigned to a fingerprint that has never been seen before.
const trustScoreNew = 10

// trustScoreKnownOtherUser is assigned when the fingerprint exists but was
// previously linked to a different user — potentially suspicious.
const trustScoreKnownOtherUser = 50

// trustScoreEstablished is assigned when the fingerprint has been seen 5 or
// more times for the requesting user.
const trustScoreEstablished = 100

// trustScoreReturning is assigned on visits 2-4 for the requesting user.
const trustScoreReturning = 75

// Servicer is the business-logic interface consumed by the HTTP handler.
type Servicer interface {
	Identify(ctx context.Context, req *domain.FingerprintRequest) (*domain.FingerprintResponse, error)
	GetByID(ctx context.Context, id string) (*domain.Fingerprint, error)
	GetUserFingerprints(ctx context.Context, userID string) ([]*domain.Fingerprint, error)
}

// fingerprintService is the concrete implementation of Servicer.
type fingerprintService struct {
	store store.FingerprintStore
}

// New constructs a fingerprintService backed by the given store.
func New(s store.FingerprintStore) Servicer {
	return &fingerprintService{store: s}
}

// Identify generates or retrieves the fingerprint for the supplied device
// attributes and user, returning a FingerprintResponse with trust scoring.
func (svc *fingerprintService) Identify(ctx context.Context, req *domain.FingerprintRequest) (*domain.FingerprintResponse, error) {
	hash := svc.hashAttributes(req.Attributes)

	existing, err := svc.store.Get(ctx, hash)
	if err != nil && err != domain.ErrNotFound {
		return nil, fmt.Errorf("service.Identify get: %w", err)
	}

	if err == domain.ErrNotFound {
		// Brand-new fingerprint — create and persist it.
		fp := &domain.Fingerprint{
			ID:         uuid.NewString(),
			Hash:       hash,
			UserID:     req.UserID,
			Attributes: req.Attributes,
			TrustScore: trustScoreNew,
			SeenCount:  1,
			FirstSeen:  time.Now().UTC(),
			LastSeen:   time.Now().UTC(),
		}

		if err = svc.store.Save(ctx, fp); err != nil {
			return nil, fmt.Errorf("service.Identify save: %w", err)
		}
		if req.UserID != "" {
			if err = svc.store.LinkToUser(ctx, req.UserID, hash); err != nil {
				return nil, fmt.Errorf("service.Identify link: %w", err)
			}
		}

		return &domain.FingerprintResponse{
			FingerprintID: fp.ID,
			Hash:          hash,
			IsKnown:       false,
			TrustScore:    trustScoreNew,
		}, nil
	}

	// Fingerprint already exists — update seen metrics.
	if err = svc.store.IncrSeen(ctx, hash); err != nil {
		return nil, fmt.Errorf("service.Identify incr: %w", err)
	}

	// Re-fetch to get the updated SeenCount for accurate scoring.
	existing, err = svc.store.Get(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("service.Identify re-fetch: %w", err)
	}

	isKnown := false
	trustScore := trustScoreNew

	if req.UserID != "" {
		userFPs, err := svc.store.GetUserFingerprints(ctx, req.UserID)
		if err != nil {
			return nil, fmt.Errorf("service.Identify user fps: %w", err)
		}

		for _, ufp := range userFPs {
			if ufp.Hash == hash {
				isKnown = true
				break
			}
		}

		if isKnown {
			// Determine trust by how many times this user has seen this device.
			switch {
			case existing.SeenCount >= 5:
				trustScore = trustScoreEstablished
			default:
				trustScore = trustScoreReturning
			}

			// Ensure it stays linked (idempotent SADD).
			_ = svc.store.LinkToUser(ctx, req.UserID, hash)
		} else {
			// The fingerprint is known but not for this user.
			if existing.UserID != "" && existing.UserID != req.UserID {
				trustScore = trustScoreKnownOtherUser
			} else {
				// First time seeing this user on this device — link it.
				trustScore = trustScoreNew
				_ = svc.store.LinkToUser(ctx, req.UserID, hash)
			}
		}
	}

	return &domain.FingerprintResponse{
		FingerprintID: existing.ID,
		Hash:          hash,
		IsKnown:       isKnown,
		TrustScore:    trustScore,
	}, nil
}

// GetByID returns the full Fingerprint record for a given UUID.
func (svc *fingerprintService) GetByID(ctx context.Context, id string) (*domain.Fingerprint, error) {
	fp, err := svc.store.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return fp, nil
}

// GetUserFingerprints returns all fingerprints associated with a user.
func (svc *fingerprintService) GetUserFingerprints(ctx context.Context, userID string) ([]*domain.Fingerprint, error) {
	fps, err := svc.store.GetUserFingerprints(ctx, userID)
	if err != nil {
		return nil, err
	}
	return fps, nil
}

// hashAttributes produces a deterministic SHA-256 hex string from the device
// attributes. Fields are sorted alphabetically before hashing so that
// attribute insertion order does not affect the result.
func (svc *fingerprintService) hashAttributes(attrs domain.DeviceAttributes) string {
	parts := []string{
		"ua=" + attrs.UserAgent,
		"lang=" + attrs.AcceptLang,
		"tz=" + attrs.Timezone,
		"res=" + attrs.ScreenRes,
		"plat=" + attrs.Platform,
		"ip=" + attrs.IPAddress,
	}
	// Include optional fields only when present so that omitting them does not
	// create a different hash from an earlier record that did not supply them.
	if attrs.Plugins != "" {
		parts = append(parts, "plug="+attrs.Plugins)
	}
	if attrs.Canvas != "" {
		parts = append(parts, "canvas="+attrs.Canvas)
	}

	sort.Strings(parts)
	raw := strings.Join(parts, "|")
	sum := sha256.Sum256([]byte(raw))
	return fmt.Sprintf("%x", sum)
}
