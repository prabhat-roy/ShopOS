// Package sender simulates FCM/APNs push notification delivery.
package sender

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/shopos/push-notification-service/internal/domain"
)

const defaultSuccessRate = 0.92

// Sender simulates push delivery to FCM (Android/Web) and APNs (iOS).
type Sender struct {
	successRate float64
	rng         *rand.Rand
}

// New creates a Sender with the default 92 % success rate.
func New() *Sender {
	return &Sender{
		successRate: defaultSuccessRate,
		rng:         rand.New(rand.NewSource(time.Now().UnixNano())), //nolint:gosec
	}
}

// NewWithRate creates a Sender with a custom success rate (0.0–1.0).
// Useful for deterministic testing.
func NewWithRate(rate float64, src rand.Source) *Sender {
	return &Sender{
		successRate: rate,
		rng:         rand.New(src), //nolint:gosec
	}
}

// Send validates the message, simulates platform-specific delivery, and
// returns a PushRecord.
func (s *Sender) Send(msg domain.PushMessage) domain.PushRecord {
	now := time.Now().UTC()

	// --- Validation ---
	if strings.TrimSpace(msg.DeviceToken) == "" {
		errMsg := "deviceToken must not be empty"
		log.Printf("[sender] FAILED messageId=%s reason=%s", msg.MessageID, errMsg)
		return domain.PushRecord{
			MessageID:   msg.MessageID,
			DeviceToken: msg.DeviceToken,
			Platform:    msg.Platform,
			Title:       msg.Title,
			Status:      "failed",
			SentAt:      now,
			ErrorMsg:    errMsg,
		}
	}

	platform := strings.ToLower(strings.TrimSpace(msg.Platform))
	if platform != domain.PlatformIOS && platform != domain.PlatformAndroid && platform != domain.PlatformWeb {
		errMsg := fmt.Sprintf("unsupported platform %q; must be ios, android, or web", msg.Platform)
		log.Printf("[sender] FAILED messageId=%s reason=%s", msg.MessageID, errMsg)
		return domain.PushRecord{
			MessageID:   msg.MessageID,
			DeviceToken: msg.DeviceToken,
			Platform:    msg.Platform,
			Title:       msg.Title,
			Status:      "failed",
			SentAt:      now,
			ErrorMsg:    errMsg,
		}
	}

	// --- Simulate delivery ---
	if s.rng.Float64() < s.successRate {
		s.logSuccess(msg, platform)
		return domain.PushRecord{
			MessageID:   msg.MessageID,
			DeviceToken: msg.DeviceToken,
			Platform:    platform,
			Title:       msg.Title,
			Status:      "delivered",
			SentAt:      now,
		}
	}

	errMsg := s.simulatedError(platform)
	log.Printf("[sender] FAILED messageId=%s platform=%s reason=%s", msg.MessageID, platform, errMsg)
	return domain.PushRecord{
		MessageID:   msg.MessageID,
		DeviceToken: msg.DeviceToken,
		Platform:    platform,
		Title:       msg.Title,
		Status:      "failed",
		SentAt:      now,
		ErrorMsg:    errMsg,
	}
}

// -------------------------------------------------------------------
// Helpers
// -------------------------------------------------------------------

func (s *Sender) logSuccess(msg domain.PushMessage, platform string) {
	switch platform {
	case domain.PlatformIOS:
		log.Printf("[sender][APNs] DELIVERED messageId=%s deviceToken=%.8s... title=%q",
			msg.MessageID, msg.DeviceToken, msg.Title)
	case domain.PlatformAndroid:
		log.Printf("[sender][FCM] DELIVERED messageId=%s deviceToken=%.8s... title=%q",
			msg.MessageID, msg.DeviceToken, msg.Title)
	case domain.PlatformWeb:
		log.Printf("[sender][WebPush] DELIVERED messageId=%s deviceToken=%.8s... title=%q",
			msg.MessageID, msg.DeviceToken, msg.Title)
	}
}

func (s *Sender) simulatedError(platform string) string {
	switch platform {
	case domain.PlatformIOS:
		return "APNs error: BadDeviceToken (410)"
	case domain.PlatformAndroid:
		return "FCM error: registration-token-not-registered"
	default:
		return "WebPush error: 410 Gone — subscription expired"
	}
}
