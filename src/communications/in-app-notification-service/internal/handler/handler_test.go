package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shopos/in-app-notification-service/internal/domain"
	"github.com/shopos/in-app-notification-service/internal/handler"
)

// mockService is a test double that satisfies service.Servicer.
type mockService struct {
	sendFn            func(ctx context.Context, userID string, t domain.NotifType, title, body, link string) (domain.Notification, error)
	getFn             func(ctx context.Context, userID string, unreadOnly bool, limit, offset int) (domain.NotifPage, error)
	markReadFn        func(ctx context.Context, userID, notifID string) error
	markAllReadFn     func(ctx context.Context, userID string) error
	unreadCountFn     func(ctx context.Context, userID string) (int, error)
	deleteFn          func(ctx context.Context, userID, notifID string) error
	clearFn           func(ctx context.Context, userID string) error
}

func (m *mockService) SendNotification(ctx context.Context, userID string, t domain.NotifType, title, body, link string) (domain.Notification, error) {
	return m.sendFn(ctx, userID, t, title, body, link)
}
func (m *mockService) GetNotifications(ctx context.Context, userID string, unreadOnly bool, limit, offset int) (domain.NotifPage, error) {
	return m.getFn(ctx, userID, unreadOnly, limit, offset)
}
func (m *mockService) MarkAsRead(ctx context.Context, userID, notifID string) error {
	return m.markReadFn(ctx, userID, notifID)
}
func (m *mockService) MarkAllAsRead(ctx context.Context, userID string) error {
	return m.markAllReadFn(ctx, userID)
}
func (m *mockService) GetUnreadCount(ctx context.Context, userID string) (int, error) {
	return m.unreadCountFn(ctx, userID)
}
func (m *mockService) DeleteNotification(ctx context.Context, userID, notifID string) error {
	return m.deleteFn(ctx, userID, notifID)
}
func (m *mockService) ClearNotifications(ctx context.Context, userID string) error {
	return m.clearFn(ctx, userID)
}

func newTestServer(svc *mockService) *httptest.Server {
	mux := http.NewServeMux()
	h := handler.New(svc)
	h.RegisterRoutes(mux)
	return httptest.NewServer(mux)
}

// Test 1: healthz returns 200 with {"status":"ok"}.
func TestHealthz(t *testing.T) {
	svc := &mockService{}
	srv := newTestServer(svc)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["status"] != "ok" {
		t.Errorf("expected status=ok, got %s", body["status"])
	}
}

// Test 2: POST /notifications/{userId} with valid payload returns 201.
func TestSendNotification_Created(t *testing.T) {
	notif := domain.Notification{ID: "n1", UserID: "u1", Type: domain.NotifTypeSystem, Title: "Hi", Body: "Body"}
	svc := &mockService{
		sendFn: func(_ context.Context, _ string, _ domain.NotifType, _, _, _ string) (domain.Notification, error) {
			return notif, nil
		},
	}
	srv := newTestServer(svc)
	defer srv.Close()

	payload := `{"type":"SYSTEM","title":"Hi","body":"Body"}`
	resp, err := http.Post(srv.URL+"/notifications/u1", "application/json", bytes.NewBufferString(payload))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected 201, got %d", resp.StatusCode)
	}
	var got domain.Notification
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.ID != "n1" {
		t.Errorf("expected id=n1, got %s", got.ID)
	}
}

// Test 3: POST /notifications/{userId} with missing title returns 400.
func TestSendNotification_BadRequest(t *testing.T) {
	svc := &mockService{}
	srv := newTestServer(svc)
	defer srv.Close()

	payload := `{"type":"SYSTEM","body":"Body"}`
	resp, err := http.Post(srv.URL+"/notifications/u1", "application/json", bytes.NewBufferString(payload))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

// Test 4: GET /notifications/{userId} returns 200 with page.
func TestGetNotifications(t *testing.T) {
	page := domain.NotifPage{
		Notifications: []domain.Notification{{ID: "n1"}},
		Total:         1,
		Unread:        1,
	}
	svc := &mockService{
		getFn: func(_ context.Context, _ string, _ bool, _, _ int) (domain.NotifPage, error) {
			return page, nil
		},
	}
	srv := newTestServer(svc)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/notifications/u1")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	var got domain.NotifPage
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.Total != 1 {
		t.Errorf("expected total=1, got %d", got.Total)
	}
}

// Test 5: GET /notifications/{userId}?unreadOnly=true passes unreadOnly=true to service.
func TestGetNotifications_UnreadOnly(t *testing.T) {
	gotUnreadOnly := false
	svc := &mockService{
		getFn: func(_ context.Context, _ string, unreadOnly bool, _, _ int) (domain.NotifPage, error) {
			gotUnreadOnly = unreadOnly
			return domain.NotifPage{Notifications: []domain.Notification{}}, nil
		},
	}
	srv := newTestServer(svc)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/notifications/u1?unreadOnly=true")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if !gotUnreadOnly {
		t.Error("expected unreadOnly=true to be forwarded to service")
	}
}

// Test 6: POST /notifications/{userId}/read-all returns 204.
func TestMarkAllRead(t *testing.T) {
	svc := &mockService{
		markAllReadFn: func(_ context.Context, _ string) error { return nil },
	}
	srv := newTestServer(svc)
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/notifications/u1/read-all", "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("expected 204, got %d", resp.StatusCode)
	}
}

// Test 7: PATCH /notifications/{userId}/{notifId}/read returns 204.
func TestMarkRead(t *testing.T) {
	svc := &mockService{
		markReadFn: func(_ context.Context, _, _ string) error { return nil },
	}
	srv := newTestServer(svc)
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodPatch, srv.URL+"/notifications/u1/n1/read", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("expected 204, got %d", resp.StatusCode)
	}
}

// Test 8: DELETE /notifications/{userId} returns 204.
func TestClearAll(t *testing.T) {
	svc := &mockService{
		clearFn: func(_ context.Context, _ string) error { return nil },
	}
	srv := newTestServer(svc)
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodDelete, srv.URL+"/notifications/u1", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("expected 204, got %d", resp.StatusCode)
	}
}

// Test 9: GET /notifications/{userId}/count returns unread count.
func TestGetUnreadCount(t *testing.T) {
	svc := &mockService{
		unreadCountFn: func(_ context.Context, _ string) (int, error) { return 7, nil },
	}
	srv := newTestServer(svc)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/notifications/u1/count")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	var body map[string]int
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["unread"] != 7 {
		t.Errorf("expected unread=7, got %d", body["unread"])
	}
}

// Test 10: DELETE /notifications/{userId}/{notifId} returns 204.
func TestDeleteNotificationByID(t *testing.T) {
	svc := &mockService{
		deleteFn: func(_ context.Context, _, _ string) error { return nil },
	}
	srv := newTestServer(svc)
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodDelete, srv.URL+"/notifications/u1/n1", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("expected 204, got %d", resp.StatusCode)
	}
}
