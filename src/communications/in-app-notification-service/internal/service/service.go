package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopos/in-app-notification-service/internal/config"
	"github.com/shopos/in-app-notification-service/internal/domain"
	"github.com/shopos/in-app-notification-service/internal/store"
)

// Servicer defines the business-logic operations for notifications.
type Servicer interface {
	SendNotification(ctx context.Context, userID string, notifType domain.NotifType, title, body, link string) (domain.Notification, error)
	GetNotifications(ctx context.Context, userID string, unreadOnly bool, limit, offset int) (domain.NotifPage, error)
	MarkAsRead(ctx context.Context, userID, notifID string) error
	MarkAllAsRead(ctx context.Context, userID string) error
	GetUnreadCount(ctx context.Context, userID string) (int, error)
	DeleteNotification(ctx context.Context, userID, notifID string) error
	ClearNotifications(ctx context.Context, userID string) error
}

// Service is the concrete implementation of Servicer.
type Service struct {
	store store.Storer
	cfg   *config.Config
}

// New creates a new Service.
func New(s store.Storer, cfg *config.Config) *Service {
	return &Service{store: s, cfg: cfg}
}

// SendNotification creates and stores a new notification for the given user.
func (s *Service) SendNotification(ctx context.Context, userID string, notifType domain.NotifType, title, body, link string) (domain.Notification, error) {
	if userID == "" {
		return domain.Notification{}, fmt.Errorf("userID is required")
	}
	n := domain.Notification{
		ID:        uuid.NewString(),
		UserID:    userID,
		Type:      notifType,
		Title:     title,
		Body:      body,
		Link:      link,
		Read:      false,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.store.SaveNotification(ctx, userID, n, s.cfg.NotifTTL, s.cfg.MaxPerUser); err != nil {
		return domain.Notification{}, fmt.Errorf("save notification: %w", err)
	}
	return n, nil
}

// GetNotifications returns a paginated page of notifications.
func (s *Service) GetNotifications(ctx context.Context, userID string, unreadOnly bool, limit, offset int) (domain.NotifPage, error) {
	if limit <= 0 {
		limit = 20
	}
	// Fetch all to calculate total and unread counts.
	all, err := s.store.GetNotifications(ctx, userID, true, 0, 0)
	if err != nil {
		return domain.NotifPage{}, fmt.Errorf("get all notifications: %w", err)
	}
	total := len(all)
	unread := 0
	for _, n := range all {
		if !n.Read {
			unread++
		}
	}

	// Fetch the requested page.
	notifs, err := s.store.GetNotifications(ctx, userID, !unreadOnly, limit, offset)
	if err != nil {
		return domain.NotifPage{}, fmt.Errorf("get notifications page: %w", err)
	}
	if notifs == nil {
		notifs = []domain.Notification{}
	}
	return domain.NotifPage{
		Notifications: notifs,
		Total:         total,
		Unread:        unread,
	}, nil
}

// MarkAsRead marks a single notification as read.
func (s *Service) MarkAsRead(ctx context.Context, userID, notifID string) error {
	if err := s.store.MarkRead(ctx, userID, notifID); err != nil {
		return fmt.Errorf("mark read: %w", err)
	}
	return nil
}

// MarkAllAsRead marks all notifications for a user as read.
func (s *Service) MarkAllAsRead(ctx context.Context, userID string) error {
	if err := s.store.MarkAllRead(ctx, userID); err != nil {
		return fmt.Errorf("mark all read: %w", err)
	}
	return nil
}

// GetUnreadCount returns the count of unread notifications.
func (s *Service) GetUnreadCount(ctx context.Context, userID string) (int, error) {
	count, err := s.store.GetUnreadCount(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("get unread count: %w", err)
	}
	return count, nil
}

// DeleteNotification deletes a single notification.
func (s *Service) DeleteNotification(ctx context.Context, userID, notifID string) error {
	if err := s.store.DeleteNotification(ctx, userID, notifID); err != nil {
		return fmt.Errorf("delete notification: %w", err)
	}
	return nil
}

// ClearNotifications removes all notifications for the user.
func (s *Service) ClearNotifications(ctx context.Context, userID string) error {
	if err := s.store.ClearAll(ctx, userID); err != nil {
		return fmt.Errorf("clear notifications: %w", err)
	}
	return nil
}
