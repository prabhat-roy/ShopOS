package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// ErrNotFound is returned when a translation key is not found.
var ErrNotFound = errors.New("translation not found")

// SupportedLocales lists all locales the service understands.
var SupportedLocales = []string{
	"en", "es", "fr", "de", "ja", "zh", "ar", "pt", "ko", "it",
}

// Translation represents a single translated string.
type Translation struct {
	ID        uuid.UUID `json:"id"`
	Locale    string    `json:"locale"`
	Namespace string    `json:"namespace"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// TranslationKey uniquely identifies a translation within a namespace.
type TranslationKey struct {
	Namespace string `json:"namespace"`
	Key       string `json:"key"`
}

// BulkUpsertRequest carries many translations at once for a given locale and namespace.
type BulkUpsertRequest struct {
	Locale       string            `json:"locale"`
	Namespace    string            `json:"namespace"`
	Translations map[string]string `json:"translations"`
}

// IsLocaleSupported returns true if the given locale is in the supported list.
func IsLocaleSupported(locale string) bool {
	for _, l := range SupportedLocales {
		if l == locale {
			return true
		}
	}
	return false
}
