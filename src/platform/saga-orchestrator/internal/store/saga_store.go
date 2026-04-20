package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/shopos/saga-orchestrator/internal/domain"
)

// SagaStore handles Postgres persistence for saga instances.
type SagaStore struct {
	db *sql.DB
}

func New(db *sql.DB) *SagaStore { return &SagaStore{db: db} }

func (s *SagaStore) Create(ctx context.Context, saga *domain.Saga) error {
	stepsJSON, _ := json.Marshal(saga.Steps)
	payloadJSON, _ := json.Marshal(saga.Payload)

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO sagas (id, type, order_id, state, steps, payload, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		saga.ID, string(saga.Type), saga.OrderID, string(saga.State),
		stepsJSON, payloadJSON, saga.CreatedAt, saga.UpdatedAt,
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return fmt.Errorf("saga %s already exists", saga.ID)
		}
		return fmt.Errorf("insert saga: %w", err)
	}
	return nil
}

func (s *SagaStore) GetByID(ctx context.Context, id string) (*domain.Saga, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, type, order_id, state, steps, payload, created_at, updated_at,
		       completed_at, failed_at, error_msg
		FROM sagas WHERE id = $1`, id)
	return scanSaga(row)
}

func (s *SagaStore) GetByOrderID(ctx context.Context, orderID string) (*domain.Saga, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, type, order_id, state, steps, payload, created_at, updated_at,
		       completed_at, failed_at, error_msg
		FROM sagas WHERE order_id = $1 ORDER BY created_at DESC LIMIT 1`, orderID)
	return scanSaga(row)
}

func (s *SagaStore) UpdateState(ctx context.Context, id string, state domain.SagaState, steps []domain.Step, errMsg string, completedAt, failedAt *time.Time) error {
	stepsJSON, _ := json.Marshal(steps)
	_, err := s.db.ExecContext(ctx, `
		UPDATE sagas
		SET state=$2, steps=$3, error_msg=$4, completed_at=$5, failed_at=$6, updated_at=NOW()
		WHERE id=$1`,
		id, string(state), stepsJSON, errMsg, completedAt, failedAt,
	)
	return err
}

func (s *SagaStore) ListByState(ctx context.Context, state domain.SagaState, limit int) ([]*domain.Saga, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, type, order_id, state, steps, payload, created_at, updated_at,
		       completed_at, failed_at, error_msg
		FROM sagas WHERE state=$1 ORDER BY created_at ASC LIMIT $2`, string(state), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sagas []*domain.Saga
	for rows.Next() {
		saga, err := scanSagaRows(rows)
		if err != nil {
			return nil, err
		}
		sagas = append(sagas, saga)
	}
	return sagas, rows.Err()
}

func scanSaga(row *sql.Row) (*domain.Saga, error) {
	var (
		saga        domain.Saga
		stepsJSON   []byte
		payloadJSON []byte
		sagaType    string
		state       string
	)
	err := row.Scan(
		&saga.ID, &sagaType, &saga.OrderID, &state,
		&stepsJSON, &payloadJSON,
		&saga.CreatedAt, &saga.UpdatedAt,
		&saga.CompletedAt, &saga.FailedAt, &saga.ErrorMsg,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	saga.Type = domain.SagaType(sagaType)
	saga.State = domain.SagaState(state)
	json.Unmarshal(stepsJSON, &saga.Steps)
	json.Unmarshal(payloadJSON, &saga.Payload)
	return &saga, nil
}

func scanSagaRows(rows *sql.Rows) (*domain.Saga, error) {
	var (
		saga        domain.Saga
		stepsJSON   []byte
		payloadJSON []byte
		sagaType    string
		state       string
	)
	err := rows.Scan(
		&saga.ID, &sagaType, &saga.OrderID, &state,
		&stepsJSON, &payloadJSON,
		&saga.CreatedAt, &saga.UpdatedAt,
		&saga.CompletedAt, &saga.FailedAt, &saga.ErrorMsg,
	)
	if err != nil {
		return nil, err
	}
	saga.Type = domain.SagaType(sagaType)
	saga.State = domain.SagaState(state)
	json.Unmarshal(stepsJSON, &saga.Steps)
	json.Unmarshal(payloadJSON, &saga.Payload)
	return &saga, nil
}
