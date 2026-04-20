package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/shopos/event-replay-service/internal/domain"
)

// Storer defines the persistence operations required by the service layer.
type Storer interface {
	Create(ctx context.Context, job *domain.ReplayJob) error
	Get(ctx context.Context, id string) (*domain.ReplayJob, error)
	List(ctx context.Context) ([]*domain.ReplayJob, error)
	UpdateStatus(ctx context.Context, id string, status domain.ReplayStatus, eventsReplayed int64, errMsg string) error
	Cancel(ctx context.Context, id string) error
}

// ReplayService implements the business logic for managing replay jobs.
type ReplayService struct {
	store         Storer
	eventStoreURL string
	client        *http.Client
}

// New constructs a ReplayService.
func New(store Storer, eventStoreURL string) *ReplayService {
	return &ReplayService{
		store:         store,
		eventStoreURL: eventStoreURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CreateJob validates and persists a new replay job in the pending state.
func (s *ReplayService) CreateJob(ctx context.Context, req *domain.CreateReplayRequest) (*domain.ReplayJob, error) {
	if req.Target == "" {
		req.Target = domain.TargetHTTP
	}

	job := &domain.ReplayJob{
		ID:             uuid.NewString(),
		StreamID:       req.StreamID,
		StreamType:     req.StreamType,
		EventType:      req.EventType,
		FromSeq:        req.FromSeq,
		ToSeq:          req.ToSeq,
		Target:         req.Target,
		TargetTopic:    req.TargetTopic,
		Status:         domain.StatusPending,
		EventsReplayed: 0,
		CreatedAt:      time.Now().UTC(),
	}

	if err := s.store.Create(ctx, job); err != nil {
		return nil, fmt.Errorf("create replay job: %w", err)
	}
	return job, nil
}

// GetJob retrieves a single replay job.
func (s *ReplayService) GetJob(ctx context.Context, id string) (*domain.ReplayJob, error) {
	return s.store.Get(ctx, id)
}

// ListJobs returns all replay jobs.
func (s *ReplayService) ListJobs(ctx context.Context) ([]*domain.ReplayJob, error) {
	return s.store.List(ctx)
}

// CancelJob cancels a pending or running replay job.
func (s *ReplayService) CancelJob(ctx context.Context, id string) error {
	return s.store.Cancel(ctx, id)
}

// RunJob executes the replay by fetching events from the event-store and
// forwarding them to the configured target. Status updates are persisted after
// each significant state change.
func (s *ReplayService) RunJob(ctx context.Context, id string) error {
	job, err := s.store.Get(ctx, id)
	if err != nil {
		return err
	}

	// Mark job as running.
	if err := s.store.UpdateStatus(ctx, id, domain.StatusRunning, 0, ""); err != nil {
		return fmt.Errorf("mark running: %w", err)
	}

	events, fetchErr := s.fetchEvents(ctx, job)
	if fetchErr != nil {
		_ = s.store.UpdateStatus(ctx, id, domain.StatusFailed, 0, fetchErr.Error())
		return fmt.Errorf("fetch events: %w", fetchErr)
	}

	var replayed int64
	for _, ev := range events {
		if sendErr := s.dispatchEvent(ctx, job, ev); sendErr != nil {
			_ = s.store.UpdateStatus(ctx, id, domain.StatusFailed, replayed, sendErr.Error())
			return fmt.Errorf("dispatch event: %w", sendErr)
		}
		replayed++
	}

	return s.store.UpdateStatus(ctx, id, domain.StatusCompleted, replayed, "")
}

// ---- private helpers --------------------------------------------------------

// eventStoreResponse is the expected envelope from the event-store service.
type eventStoreResponse struct {
	Events []json.RawMessage `json:"events"`
}

// fetchEvents calls the event-store service to retrieve the events that should
// be replayed for the given job.
func (s *ReplayService) fetchEvents(ctx context.Context, job *domain.ReplayJob) ([]json.RawMessage, error) {
	url := fmt.Sprintf("%s/streams/%s/events?from_seq=%d&to_seq=%d",
		s.eventStoreURL, job.StreamID, job.FromSeq, job.ToSeq)

	if job.EventType != "" {
		url += fmt.Sprintf("&event_type=%s", job.EventType)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("event-store returned HTTP %d", resp.StatusCode)
	}

	var body eventStoreResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return body.Events, nil
}

// dispatchEvent forwards a single event payload to the job's target endpoint.
func (s *ReplayService) dispatchEvent(ctx context.Context, job *domain.ReplayJob, event json.RawMessage) error {
	var targetURL string
	switch job.Target {
	case domain.TargetHTTP:
		targetURL = job.TargetTopic
		if targetURL == "" {
			// Default to a well-known replay-events endpoint on the event-store.
			targetURL = fmt.Sprintf("%s/replay-events", s.eventStoreURL)
		}
	case domain.TargetKafka:
		// When targeting Kafka the TargetTopic carries the topic name.
		// We proxy via the event-store's Kafka bridge endpoint.
		targetURL = fmt.Sprintf("%s/kafka/topics/%s/messages", s.eventStoreURL, job.TargetTopic)
	default:
		return fmt.Errorf("unknown target type: %s", job.Target)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, bytes.NewReader(event))
	if err != nil {
		return fmt.Errorf("build dispatch request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("POST %s: %w", targetURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("target returned HTTP %d", resp.StatusCode)
	}
	return nil
}
