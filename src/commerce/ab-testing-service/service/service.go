package service

import (
	"context"
	"fmt"
	"hash/fnv"
	"time"

	"github.com/google/uuid"

	"github.com/shopos/ab-testing-service/domain"
	"github.com/shopos/ab-testing-service/store"
)

// ExperimentService implements business logic for A/B experiments.
type ExperimentService struct {
	store *store.Store
}

// New creates a new ExperimentService.
func New(s *store.Store) *ExperimentService {
	return &ExperimentService{store: s}
}

// CreateExperiment persists a new experiment and returns it with a generated ID.
func (svc *ExperimentService) CreateExperiment(ctx context.Context, e *domain.Experiment) (*domain.Experiment, error) {
	if e.ID == "" {
		e.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	e.CreatedAt = now
	e.UpdatedAt = now
	if e.TrafficPercent == 0 {
		e.TrafficPercent = 100
	}
	if err := svc.store.SaveExperiment(ctx, e); err != nil {
		return nil, fmt.Errorf("save experiment: %w", err)
	}
	return e, nil
}

// GetExperiment returns an experiment by ID.
func (svc *ExperimentService) GetExperiment(ctx context.Context, id string) (*domain.Experiment, error) {
	return svc.store.GetExperiment(ctx, id)
}

// ListExperiments returns all experiments.
func (svc *ExperimentService) ListExperiments(ctx context.Context) ([]*domain.Experiment, error) {
	return svc.store.ListExperiments(ctx)
}

// Assign deterministically assigns a user to a variant using FNV hashing.
// The assignment is persisted so the same user always gets the same variant.
func (svc *ExperimentService) Assign(ctx context.Context, experimentID, userID string) (*domain.Assignment, error) {
	// Return existing assignment if already made
	existing, err := svc.store.GetAssignment(ctx, experimentID, userID)
	if err == nil {
		return existing, nil
	}
	if err != domain.ErrNotFound {
		return nil, fmt.Errorf("get assignment: %w", err)
	}

	exp, err := svc.store.GetExperiment(ctx, experimentID)
	if err != nil {
		return nil, err
	}
	if len(exp.Variants) == 0 {
		return nil, fmt.Errorf("experiment %s has no variants", experimentID)
	}

	variant := pickVariant(experimentID, userID, exp.Variants)
	a := &domain.Assignment{
		ExperimentID: experimentID,
		UserID:       userID,
		Variant:      variant,
		AssignedAt:   time.Now().UTC(),
	}
	if err := svc.store.SaveAssignment(ctx, a); err != nil {
		return nil, fmt.Errorf("save assignment: %w", err)
	}
	return a, nil
}

// GetAssignment retrieves an existing assignment.
func (svc *ExperimentService) GetAssignment(ctx context.Context, experimentID, userID string) (*domain.Assignment, error) {
	return svc.store.GetAssignment(ctx, experimentID, userID)
}

// RecordConversion saves a metric conversion event.
func (svc *ExperimentService) RecordConversion(ctx context.Context, experimentID, userID, metric string, value float64) error {
	c := &domain.Conversion{
		ExperimentID: experimentID,
		UserID:       userID,
		Metric:       metric,
		Value:        value,
		RecordedAt:   time.Now().UTC(),
	}
	return svc.store.RecordConversion(ctx, c)
}

// pickVariant deterministically selects a variant using FNV-32a hash over the
// composite key "<experimentID>:<userID>".  Variants are weighted; the hash
// bucket maps into the cumulative weight distribution.
func pickVariant(experimentID, userID string, variants []domain.Variant) string {
	h := fnv.New32a()
	h.Write([]byte(experimentID + ":" + userID)) //nolint:errcheck

	totalWeight := 0
	for _, v := range variants {
		totalWeight += v.Weight
	}
	if totalWeight == 0 {
		return variants[0].Name
	}

	bucket := int(h.Sum32()) % totalWeight
	cumulative := 0
	for _, v := range variants {
		cumulative += v.Weight
		if bucket < cumulative {
			return v.Name
		}
	}
	return variants[len(variants)-1].Name
}
