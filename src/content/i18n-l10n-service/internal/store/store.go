package store

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/shopos/i18n-l10n-service/internal/domain"
)

// Storer defines the data-access contract.
type Storer interface {
	GetTranslation(locale, namespace, key string) (string, error)
	GetNamespace(locale, namespace string) (map[string]string, error)
	UpsertTranslation(t domain.Translation) error
	BulkUpsert(locale, namespace string, translations map[string]string) error
	DeleteTranslation(locale, namespace, key string) error
	ListLocales() ([]string, error)
	ListNamespaces(locale string) ([]string, error)
	GetWithFallback(locale, namespace, key, fallbackLocale string) (value string, resolvedLocale string, err error)
	Ping() error
	Close() error
}

// PostgresStore implements Storer against a PostgreSQL database.
type PostgresStore struct {
	db *sql.DB
}

// New opens a PostgreSQL connection and verifies it with a ping.
func New(dsn string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("store: open db: %w", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	return &PostgresStore{db: db}, nil
}

func (s *PostgresStore) Ping() error {
	return s.db.Ping()
}

func (s *PostgresStore) Close() error {
	return s.db.Close()
}

// GetTranslation retrieves a single translation value.
func (s *PostgresStore) GetTranslation(locale, namespace, key string) (string, error) {
	var value string
	err := s.db.QueryRow(
		`SELECT value FROM translations WHERE locale=$1 AND namespace=$2 AND key=$3`,
		locale, namespace, key,
	).Scan(&value)
	if err == sql.ErrNoRows {
		return "", domain.ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("store: get translation: %w", err)
	}
	return value, nil
}

// GetNamespace returns all key-value pairs for a locale+namespace.
func (s *PostgresStore) GetNamespace(locale, namespace string) (map[string]string, error) {
	rows, err := s.db.Query(
		`SELECT key, value FROM translations WHERE locale=$1 AND namespace=$2 ORDER BY key`,
		locale, namespace,
	)
	if err != nil {
		return nil, fmt.Errorf("store: get namespace: %w", err)
	}
	defer rows.Close()

	result := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, fmt.Errorf("store: scan row: %w", err)
		}
		result[k] = v
	}
	return result, rows.Err()
}

// UpsertTranslation inserts or updates a single translation.
func (s *PostgresStore) UpsertTranslation(t domain.Translation) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	now := time.Now().UTC()
	_, err := s.db.Exec(
		`INSERT INTO translations (id, locale, namespace, key, value, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 ON CONFLICT (locale, namespace, key)
		 DO UPDATE SET value=EXCLUDED.value, updated_at=EXCLUDED.updated_at`,
		t.ID, t.Locale, t.Namespace, t.Key, t.Value, now, now,
	)
	if err != nil {
		return fmt.Errorf("store: upsert translation: %w", err)
	}
	return nil
}

// BulkUpsert inserts or updates many translations in a single transaction.
func (s *PostgresStore) BulkUpsert(locale, namespace string, translations map[string]string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("store: bulk upsert begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	stmt, err := tx.Prepare(
		`INSERT INTO translations (id, locale, namespace, key, value, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 ON CONFLICT (locale, namespace, key)
		 DO UPDATE SET value=EXCLUDED.value, updated_at=EXCLUDED.updated_at`,
	)
	if err != nil {
		return fmt.Errorf("store: bulk upsert prepare: %w", err)
	}
	defer stmt.Close()

	now := time.Now().UTC()
	for k, v := range translations {
		if _, err = stmt.Exec(uuid.New(), locale, namespace, k, v, now, now); err != nil {
			return fmt.Errorf("store: bulk upsert exec key=%s: %w", k, err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("store: bulk upsert commit: %w", err)
	}
	return nil
}

// DeleteTranslation removes a translation entry.
func (s *PostgresStore) DeleteTranslation(locale, namespace, key string) error {
	res, err := s.db.Exec(
		`DELETE FROM translations WHERE locale=$1 AND namespace=$2 AND key=$3`,
		locale, namespace, key,
	)
	if err != nil {
		return fmt.Errorf("store: delete translation: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// ListLocales returns all distinct locales stored in the database.
func (s *PostgresStore) ListLocales() ([]string, error) {
	rows, err := s.db.Query(`SELECT DISTINCT locale FROM translations ORDER BY locale`)
	if err != nil {
		return nil, fmt.Errorf("store: list locales: %w", err)
	}
	defer rows.Close()

	var locales []string
	for rows.Next() {
		var l string
		if err := rows.Scan(&l); err != nil {
			return nil, err
		}
		locales = append(locales, l)
	}
	return locales, rows.Err()
}

// ListNamespaces returns all distinct namespaces for a given locale.
func (s *PostgresStore) ListNamespaces(locale string) ([]string, error) {
	rows, err := s.db.Query(
		`SELECT DISTINCT namespace FROM translations WHERE locale=$1 ORDER BY namespace`,
		locale,
	)
	if err != nil {
		return nil, fmt.Errorf("store: list namespaces: %w", err)
	}
	defer rows.Close()

	var namespaces []string
	for rows.Next() {
		var ns string
		if err := rows.Scan(&ns); err != nil {
			return nil, err
		}
		namespaces = append(namespaces, ns)
	}
	return namespaces, rows.Err()
}

// GetWithFallback tries locale first; if not found, tries fallbackLocale.
// Returns the value and the locale it was resolved from.
func (s *PostgresStore) GetWithFallback(locale, namespace, key, fallbackLocale string) (string, string, error) {
	value, err := s.GetTranslation(locale, namespace, key)
	if err == nil {
		return value, locale, nil
	}
	if err != domain.ErrNotFound {
		return "", "", err
	}

	// Try fallback
	if fallbackLocale != "" && fallbackLocale != locale {
		value, err = s.GetTranslation(fallbackLocale, namespace, key)
		if err == nil {
			return value, fallbackLocale, nil
		}
		if err != domain.ErrNotFound {
			return "", "", err
		}
	}

	return "", "", domain.ErrNotFound
}
