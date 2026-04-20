package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopos/live-chat-service/internal/domain"
)

// Storer is the persistence interface required by the service layer.
type Storer interface {
	CreateSession(ctx context.Context, session domain.ChatSession) error
	GetSession(ctx context.Context, id string) (domain.ChatSession, error)
	UpdateSession(ctx context.Context, session domain.ChatSession) error
	SaveMessage(ctx context.Context, msg domain.ChatMessage) error
	GetMessages(ctx context.Context, sessionID string, limit int) ([]domain.ChatMessage, error)
	ListWaitingSessions(ctx context.Context) ([]domain.ChatSession, error)
}

// Service implements business logic for live chat.
type Service struct {
	store       Storer
	maxMessages int
}

// New creates a Service with the provided store and message limit.
func New(store Storer, maxMessages int) *Service {
	return &Service{store: store, maxMessages: maxMessages}
}

// StartSession creates a new chat session in "waiting" status for the given customer.
func (s *Service) StartSession(ctx context.Context, customerID string) (domain.ChatSession, error) {
	if customerID == "" {
		return domain.ChatSession{}, fmt.Errorf("customerID is required")
	}
	now := time.Now().UTC()
	session := domain.ChatSession{
		ID:         uuid.NewString(),
		CustomerID: customerID,
		Status:     domain.StatusWaiting,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := s.store.CreateSession(ctx, session); err != nil {
		return domain.ChatSession{}, fmt.Errorf("create session: %w", err)
	}
	// Emit system message
	_ = s.emitSystemMessage(ctx, session.ID, "Chat session started. Please wait for an agent.")
	return session, nil
}

// AssignAgent assigns an agent to a waiting session and transitions it to "active".
func (s *Service) AssignAgent(ctx context.Context, sessionID, agentID string) (domain.ChatSession, error) {
	session, err := s.store.GetSession(ctx, sessionID)
	if err != nil {
		return domain.ChatSession{}, err
	}
	if session.Status == domain.StatusClosed {
		return domain.ChatSession{}, domain.ErrSessionClosed
	}
	session.AgentID = agentID
	session.Status = domain.StatusActive
	session.UpdatedAt = time.Now().UTC()
	if err := s.store.UpdateSession(ctx, session); err != nil {
		return domain.ChatSession{}, fmt.Errorf("update session: %w", err)
	}
	_ = s.emitSystemMessage(ctx, sessionID, fmt.Sprintf("Agent %s has joined the chat.", agentID))
	return session, nil
}

// SendMessage adds a message to the session.
func (s *Service) SendMessage(ctx context.Context, sessionID, senderID, senderType, body string) (domain.ChatMessage, error) {
	session, err := s.store.GetSession(ctx, sessionID)
	if err != nil {
		return domain.ChatMessage{}, err
	}
	if session.Status == domain.StatusClosed {
		return domain.ChatMessage{}, domain.ErrSessionClosed
	}
	msg := domain.ChatMessage{
		ID:         uuid.NewString(),
		SessionID:  sessionID,
		SenderID:   senderID,
		SenderType: senderType,
		Body:       body,
		SentAt:     time.Now().UTC(),
	}
	if err := s.store.SaveMessage(ctx, msg); err != nil {
		return domain.ChatMessage{}, fmt.Errorf("save message: %w", err)
	}
	return msg, nil
}

// GetSession retrieves a chat session by ID.
func (s *Service) GetSession(ctx context.Context, sessionID string) (domain.ChatSession, error) {
	return s.store.GetSession(ctx, sessionID)
}

// GetMessages retrieves recent messages for a session.
func (s *Service) GetMessages(ctx context.Context, sessionID string) ([]domain.ChatMessage, error) {
	_, err := s.store.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	return s.store.GetMessages(ctx, sessionID, s.maxMessages)
}

// CloseSession closes a session and appends a system message.
func (s *Service) CloseSession(ctx context.Context, sessionID string) (domain.ChatSession, error) {
	session, err := s.store.GetSession(ctx, sessionID)
	if err != nil {
		return domain.ChatSession{}, err
	}
	if session.Status == domain.StatusClosed {
		return session, nil
	}
	session.Status = domain.StatusClosed
	session.UpdatedAt = time.Now().UTC()
	if err := s.store.UpdateSession(ctx, session); err != nil {
		return domain.ChatSession{}, fmt.Errorf("update session: %w", err)
	}
	_ = s.emitSystemMessage(ctx, sessionID, "Chat session has been closed.")
	return session, nil
}

// ListWaitingSessions returns all sessions waiting for an agent.
func (s *Service) ListWaitingSessions(ctx context.Context) ([]domain.ChatSession, error) {
	return s.store.ListWaitingSessions(ctx)
}

func (s *Service) emitSystemMessage(ctx context.Context, sessionID, text string) error {
	msg := domain.ChatMessage{
		ID:         uuid.NewString(),
		SessionID:  sessionID,
		SenderID:   "system",
		SenderType: domain.SenderSystem,
		Body:       text,
		SentAt:     time.Now().UTC(),
	}
	return s.store.SaveMessage(ctx, msg)
}
