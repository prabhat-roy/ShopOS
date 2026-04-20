package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/shopos/event-store-service/internal/domain"
)

// EventStore handles all Postgres persistence for events and snapshots.
type EventStore struct {
	db         *sql.DB
	maxPerPage int
}

func New(db *sql.DB, maxPerPage int) *EventStore {
	return &EventStore{db: db, maxPerPage: maxPerPage}
}

// Append writes events to a stream atomically, enforcing optimistic concurrency.
// ExpectedVersion -1 skips the version check; 0 means the stream must not exist.
func (s *EventStore) Append(ctx context.Context, req *domain.AppendRequest) ([]*domain.Event, error) {
	if req.StreamID == "" || req.StreamType == "" || len(req.Events) == 0 {
		return nil, domain.ErrInvalidInput
	}

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Lock the stream row and fetch current version
	var currentVersion int64 = -1
	row := tx.QueryRowContext(ctx,
		`SELECT COALESCE(MAX(version), -1) FROM events WHERE stream_id = $1`,
		req.StreamID,
	)
	if err := row.Scan(&currentVersion); err != nil {
		return nil, err
	}

	// Optimistic concurrency check
	if req.ExpectedVersion >= 0 && currentVersion != req.ExpectedVersion {
		return nil, fmt.Errorf("%w: expected %d, current %d",
			domain.ErrVersionConflict, req.ExpectedVersion, currentVersion)
	}

	now := time.Now().UTC()
	appended := make([]*domain.Event, 0, len(req.Events))

	for i, ne := range req.Events {
		version := currentVersion + int64(i) + 1
		id := uuid.New().String()
		occurredAt := now
		if ne.OccurredAt != nil {
			occurredAt = *ne.OccurredAt
		}
		meta := ne.Metadata
		if meta == nil {
			meta = []byte("{}")
		}

		var globalSeq int64
		err := tx.QueryRowContext(ctx, `
			INSERT INTO events (id, stream_id, stream_type, event_type, version, payload, metadata, occurred_at, recorded_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			RETURNING global_seq`,
			id, req.StreamID, req.StreamType, ne.EventType,
			version, ne.Payload, meta, occurredAt, now,
		).Scan(&globalSeq)
		if err != nil {
			var pqErr *pq.Error
			if errors.As(err, &pqErr) && pqErr.Code == "23505" {
				return nil, domain.ErrVersionConflict
			}
			return nil, fmt.Errorf("insert event: %w", err)
		}

		appended = append(appended, &domain.Event{
			ID:         id,
			StreamID:   req.StreamID,
			StreamType: req.StreamType,
			EventType:  ne.EventType,
			Version:    version,
			GlobalSeq:  globalSeq,
			Payload:    ne.Payload,
			Metadata:   meta,
			OccurredAt: occurredAt,
			RecordedAt: now,
		})
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return appended, nil
}

// Read returns events from a single stream ordered by version.
func (s *EventStore) Read(ctx context.Context, req domain.ReadRequest) ([]*domain.Event, error) {
	if req.StreamID == "" {
		return nil, domain.ErrInvalidInput
	}
	limit := s.clampLimit(req.MaxCount)

	query := `SELECT id, stream_id, stream_type, event_type, version, global_seq,
	                 payload, metadata, occurred_at, recorded_at
	          FROM events
	          WHERE stream_id = $1 AND version >= $2`
	args := []any{req.StreamID, req.FromVersion}

	if req.ToVersion > 0 {
		query += fmt.Sprintf(" AND version <= $%d", len(args)+1)
		args = append(args, req.ToVersion)
	}
	query += fmt.Sprintf(" ORDER BY version ASC LIMIT $%d", len(args)+1)
	args = append(args, limit)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanEvents(rows)
}

// ReadAll returns events globally ordered by global_seq with optional type filters.
func (s *EventStore) ReadAll(ctx context.Context, req domain.ReadAllRequest) ([]*domain.Event, error) {
	limit := s.clampLimit(req.MaxCount)

	query := `SELECT id, stream_id, stream_type, event_type, version, global_seq,
	                 payload, metadata, occurred_at, recorded_at
	          FROM events WHERE global_seq >= $1`
	args := []any{req.FromGlobalSeq}

	if req.StreamType != "" {
		args = append(args, req.StreamType)
		query += fmt.Sprintf(" AND stream_type = $%d", len(args))
	}
	if req.EventType != "" {
		args = append(args, req.EventType)
		query += fmt.Sprintf(" AND event_type = $%d", len(args))
	}
	query += fmt.Sprintf(" ORDER BY global_seq ASC LIMIT $%d", len(args)+1)
	args = append(args, limit)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanEvents(rows)
}

// GetByID retrieves a single event by its UUID.
func (s *EventStore) GetByID(ctx context.Context, id string) (*domain.Event, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, stream_id, stream_type, event_type, version, global_seq,
		       payload, metadata, occurred_at, recorded_at
		FROM events WHERE id = $1`, id)
	ev, err := scanEvent(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return ev, err
}

// SaveSnapshot upserts a snapshot for a stream version.
func (s *EventStore) SaveSnapshot(ctx context.Context, snap *domain.Snapshot) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO snapshots (stream_id, stream_type, version, state, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (stream_id) DO UPDATE
		SET version=$3, state=$4, created_at=$5`,
		snap.StreamID, snap.StreamType, snap.Version, snap.State, snap.CreatedAt,
	)
	return err
}

// GetSnapshot retrieves the latest snapshot for a stream.
func (s *EventStore) GetSnapshot(ctx context.Context, streamID string) (*domain.Snapshot, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT stream_id, stream_type, version, state, created_at
		FROM snapshots WHERE stream_id = $1`, streamID)
	var snap domain.Snapshot
	err := row.Scan(&snap.StreamID, &snap.StreamType, &snap.Version, &snap.State, &snap.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return &snap, err
}

func (s *EventStore) clampLimit(requested int) int {
	if requested <= 0 || requested > s.maxPerPage {
		return s.maxPerPage
	}
	return requested
}

func scanEvents(rows *sql.Rows) ([]*domain.Event, error) {
	var events []*domain.Event
	for rows.Next() {
		ev := &domain.Event{}
		if err := rows.Scan(
			&ev.ID, &ev.StreamID, &ev.StreamType, &ev.EventType,
			&ev.Version, &ev.GlobalSeq,
			&ev.Payload, &ev.Metadata, &ev.OccurredAt, &ev.RecordedAt,
		); err != nil {
			return nil, err
		}
		events = append(events, ev)
	}
	return events, rows.Err()
}

func scanEvent(row *sql.Row) (*domain.Event, error) {
	ev := &domain.Event{}
	err := row.Scan(
		&ev.ID, &ev.StreamID, &ev.StreamType, &ev.EventType,
		&ev.Version, &ev.GlobalSeq,
		&ev.Payload, &ev.Metadata, &ev.OccurredAt, &ev.RecordedAt,
	)
	return ev, err
}
