package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/shopos/mfa-service/internal/domain"
)

// MFAStore provides persistence operations for MFA setups and backup codes.
type MFAStore struct {
	db *sql.DB
}

// New opens a Postgres connection pool and returns a ready MFAStore.
func New(databaseURL string) (*MFAStore, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("store: open db: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return &MFAStore{db: db}, nil
}

// Close releases the underlying database connection pool.
func (s *MFAStore) Close() error {
	return s.db.Close()
}

// Save upserts an MFASetup record and inserts any provided backup codes.
// Backup codes are stored as SHA-256 hashes. If BackupCodes is nil/empty the
// existing codes are left untouched.
func (s *MFAStore) Save(ctx context.Context, setup *domain.MFASetup) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("store: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	now := time.Now().UTC()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO mfa_setups (user_id, secret, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id) DO UPDATE
		  SET secret     = EXCLUDED.secret,
		      status     = EXCLUDED.status,
		      updated_at = EXCLUDED.updated_at
	`, setup.UserID, setup.Secret, string(setup.Status), now, now)
	if err != nil {
		return fmt.Errorf("store: upsert mfa_setups: %w", err)
	}

	if len(setup.BackupCodes) > 0 {
		// Remove stale codes first so a re-enroll starts fresh.
		_, err = tx.ExecContext(ctx,
			`DELETE FROM mfa_backup_codes WHERE user_id = $1`, setup.UserID)
		if err != nil {
			return fmt.Errorf("store: delete old backup codes: %w", err)
		}

		for _, code := range setup.BackupCodes {
			hash := hashCode(code)
			_, err = tx.ExecContext(ctx, `
				INSERT INTO mfa_backup_codes (id, user_id, code_hash, used)
				VALUES ($1, $2, $3, FALSE)
			`, uuid.New().String(), setup.UserID, hash)
			if err != nil {
				return fmt.Errorf("store: insert backup code: %w", err)
			}
		}
	}

	return tx.Commit()
}

// Get retrieves the MFASetup for the given userID.
// Returns domain.ErrNotFound if no row exists.
func (s *MFAStore) Get(ctx context.Context, userID string) (*domain.MFASetup, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT user_id, secret, status, created_at, updated_at
		FROM   mfa_setups
		WHERE  user_id = $1
	`, userID)

	setup := &domain.MFASetup{}
	var status string
	err := row.Scan(
		&setup.UserID,
		&setup.Secret,
		&status,
		&setup.CreatedAt,
		&setup.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("store: scan mfa_setups: %w", err)
	}

	setup.Status = domain.MFAStatus(status)
	return setup, nil
}

// UpdateStatus sets the status column for the given userID.
// Returns domain.ErrNotFound if no row exists.
func (s *MFAStore) UpdateStatus(ctx context.Context, userID string, status domain.MFAStatus) error {
	res, err := s.db.ExecContext(ctx, `
		UPDATE mfa_setups
		SET    status = $1, updated_at = $2
		WHERE  user_id = $3
	`, string(status), time.Now().UTC(), userID)
	if err != nil {
		return fmt.Errorf("store: update status: %w", err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("store: rows affected: %w", err)
	}
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// UseBackupCode marks a backup code as used atomically.
// Returns domain.ErrNotFound if the code does not exist for the user or has
// already been consumed.
func (s *MFAStore) UseBackupCode(ctx context.Context, userID, code string) error {
	hash := hashCode(code)
	now := time.Now().UTC()

	res, err := s.db.ExecContext(ctx, `
		UPDATE mfa_backup_codes
		SET    used = TRUE, used_at = $1
		WHERE  user_id = $2
		  AND  code_hash = $3
		  AND  used = FALSE
	`, now, userID, hash)
	if err != nil {
		return fmt.Errorf("store: use backup code: %w", err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("store: rows affected: %w", err)
	}
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// hashCode returns the hex-encoded SHA-256 digest of a backup code.
func hashCode(code string) string {
	sum := sha256.Sum256([]byte(code))
	return fmt.Sprintf("%x", sum)
}
