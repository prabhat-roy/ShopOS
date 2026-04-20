package store

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq" // Postgres driver

	"github.com/shopos/dead-letter-service/internal/domain"
)

// DLQStore provides Postgres-backed persistence for dead-lettered messages.
type DLQStore struct {
	db *sql.DB
}

// New creates a new DLQStore using the provided *sql.DB connection.
func New(db *sql.DB) *DLQStore {
	return &DLQStore{db: db}
}

// Save inserts a new DeadMessage into the database.
func (s *DLQStore) Save(msg *domain.DeadMessage) error {
	const q = `
		INSERT INTO dead_messages
			(id, topic, partition, offset, key, payload, error_reason, status, retry_count, created_at, updated_at)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err := s.db.Exec(q,
		msg.ID,
		msg.Topic,
		msg.Partition,
		msg.Offset,
		msg.Key,
		msg.Payload,
		msg.ErrorReason,
		string(msg.Status),
		msg.RetryCount,
		msg.CreatedAt,
		msg.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("store.Save: %w", err)
	}
	return nil
}

// Get retrieves a single DeadMessage by its ID.
// Returns domain.ErrNotFound when the row does not exist.
func (s *DLQStore) Get(id string) (*domain.DeadMessage, error) {
	const q = `
		SELECT id, topic, partition, "offset", key, payload, error_reason, status, retry_count, created_at, updated_at
		FROM dead_messages
		WHERE id = $1`

	row := s.db.QueryRow(q, id)
	msg, err := scanMessage(row)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("store.Get: %w", err)
	}
	return msg, nil
}

// List retrieves a paginated list of DeadMessages filtered by topic and/or status.
// Passing empty strings for topic or status omits those filters.
func (s *DLQStore) List(topic string, status domain.MessageStatus, limit, offset int) ([]*domain.DeadMessage, error) {
	base := `
		SELECT id, topic, partition, "offset", key, payload, error_reason, status, retry_count, created_at, updated_at
		FROM dead_messages
		WHERE 1=1`

	args := []interface{}{}
	argIdx := 1

	if topic != "" {
		base += fmt.Sprintf(" AND topic = $%d", argIdx)
		args = append(args, topic)
		argIdx++
	}
	if status != "" {
		base += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, string(status))
		argIdx++
	}

	base += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := s.db.Query(base, args...)
	if err != nil {
		return nil, fmt.Errorf("store.List: %w", err)
	}
	defer rows.Close()

	var msgs []*domain.DeadMessage
	for rows.Next() {
		msg, err := scanRow(rows)
		if err != nil {
			return nil, fmt.Errorf("store.List scan: %w", err)
		}
		msgs = append(msgs, msg)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store.List rows: %w", err)
	}
	return msgs, nil
}

// UpdateStatus changes the status of a message identified by id and increments
// retry_count when the new status is StatusRetried.
// Returns domain.ErrNotFound when the row does not exist.
func (s *DLQStore) UpdateStatus(id string, status domain.MessageStatus) error {
	var q string
	if status == domain.StatusRetried {
		q = `UPDATE dead_messages
			 SET status = $1, retry_count = retry_count + 1, updated_at = $2
			 WHERE id = $3`
	} else {
		q = `UPDATE dead_messages
			 SET status = $1, updated_at = $2
			 WHERE id = $3`
	}

	res, err := s.db.Exec(q, string(status), time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("store.UpdateStatus: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("store.UpdateStatus rows affected: %w", err)
	}
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// Stats returns the count of messages grouped by status.
func (s *DLQStore) Stats() (map[string]int64, error) {
	const q = `SELECT status, COUNT(*) FROM dead_messages GROUP BY status`

	rows, err := s.db.Query(q)
	if err != nil {
		return nil, fmt.Errorf("store.Stats: %w", err)
	}
	defer rows.Close()

	result := map[string]int64{
		string(domain.StatusPending):   0,
		string(domain.StatusRetried):   0,
		string(domain.StatusDiscarded): 0,
	}

	for rows.Next() {
		var status string
		var count int64
		if err := rows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("store.Stats scan: %w", err)
		}
		result[status] = count
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store.Stats rows: %w", err)
	}
	return result, nil
}

// scanner is a common interface for *sql.Row and *sql.Rows so we can share scanMessage.
type scanner interface {
	Scan(dest ...interface{}) error
}

func scanMessage(s scanner) (*domain.DeadMessage, error) {
	var msg domain.DeadMessage
	var status string
	err := s.Scan(
		&msg.ID,
		&msg.Topic,
		&msg.Partition,
		&msg.Offset,
		&msg.Key,
		&msg.Payload,
		&msg.ErrorReason,
		&status,
		&msg.RetryCount,
		&msg.CreatedAt,
		&msg.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	msg.Status = domain.MessageStatus(status)
	return &msg, nil
}

func scanRow(rows *sql.Rows) (*domain.DeadMessage, error) {
	var msg domain.DeadMessage
	var status string
	err := rows.Scan(
		&msg.ID,
		&msg.Topic,
		&msg.Partition,
		&msg.Offset,
		&msg.Key,
		&msg.Payload,
		&msg.ErrorReason,
		&status,
		&msg.RetryCount,
		&msg.CreatedAt,
		&msg.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	msg.Status = domain.MessageStatus(status)
	return &msg, nil
}
