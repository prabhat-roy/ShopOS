package service

import (
	"fmt"

	"github.com/shopos/i18n-l10n-service/internal/domain"
	"github.com/shopos/i18n-l10n-service/internal/store"
)

// Servicer is the business-logic contract.
type Servicer interface {
	GetTranslation(locale, namespace, key string) (string, error)
	GetNamespace(locale, namespace string) (map[string]string, error)
	UpsertTranslation(t domain.Translation) error
	BulkUpsert(req domain.BulkUpsertRequest) error
	DeleteTranslation(locale, namespace, key string) error
	ListLocales() ([]string, error)
	ListNamespaces(locale string) ([]string, error)
	GetSupportedLocales() []string
}

// I18nService implements Servicer.
type I18nService struct {
	store         store.Storer
	defaultLocale string
}

// New creates a new I18nService.
func New(s store.Storer, defaultLocale string) *I18nService {
	return &I18nService{store: s, defaultLocale: defaultLocale}
}

// GetTranslation fetches a single translation with automatic fallback to defaultLocale.
func (svc *I18nService) GetTranslation(locale, namespace, key string) (string, error) {
	if locale == "" {
		locale = svc.defaultLocale
	}
	value, _, err := svc.store.GetWithFallback(locale, namespace, key, svc.defaultLocale)
	if err != nil {
		return "", fmt.Errorf("service: get translation: %w", err)
	}
	return value, nil
}

// GetNamespace returns all translations for a locale+namespace.
func (svc *I18nService) GetNamespace(locale, namespace string) (map[string]string, error) {
	if locale == "" {
		locale = svc.defaultLocale
	}
	result, err := svc.store.GetNamespace(locale, namespace)
	if err != nil {
		return nil, fmt.Errorf("service: get namespace: %w", err)
	}
	return result, nil
}

// UpsertTranslation creates or updates a single translation.
func (svc *I18nService) UpsertTranslation(t domain.Translation) error {
	if t.Locale == "" || t.Namespace == "" || t.Key == "" {
		return fmt.Errorf("service: locale, namespace and key are required")
	}
	return svc.store.UpsertTranslation(t)
}

// BulkUpsert processes a bulk upsert request.
func (svc *I18nService) BulkUpsert(req domain.BulkUpsertRequest) error {
	if req.Locale == "" || req.Namespace == "" {
		return fmt.Errorf("service: locale and namespace are required")
	}
	if len(req.Translations) == 0 {
		return fmt.Errorf("service: translations map must not be empty")
	}
	return svc.store.BulkUpsert(req.Locale, req.Namespace, req.Translations)
}

// DeleteTranslation removes a translation.
func (svc *I18nService) DeleteTranslation(locale, namespace, key string) error {
	return svc.store.DeleteTranslation(locale, namespace, key)
}

// ListLocales returns all known locales from storage.
func (svc *I18nService) ListLocales() ([]string, error) {
	return svc.store.ListLocales()
}

// ListNamespaces returns all namespaces for a locale.
func (svc *I18nService) ListNamespaces(locale string) ([]string, error) {
	return svc.store.ListNamespaces(locale)
}

// GetSupportedLocales returns the hardcoded list of supported locales.
func (svc *I18nService) GetSupportedLocales() []string {
	return domain.SupportedLocales
}
