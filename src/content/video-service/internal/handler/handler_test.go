package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/shopos/video-service/internal/domain"
	"github.com/shopos/video-service/internal/handler"
)

// ---- fake service -----------------------------------------------------------

type fakeService struct {
	videos map[string]domain.Video
}

func newFakeService() *fakeService {
	return &fakeService{videos: make(map[string]domain.Video)}
}

func (f *fakeService) UploadVideo(_ context.Context, ownerID, title, description, originalName, contentType string, _ io.Reader, size int64) (domain.Video, error) {
	now := time.Now().UTC()
	v := domain.Video{
		ID:           "vid-001",
		Title:        title,
		Description:  description,
		OriginalName: originalName,
		StoredKey:    "vid-001/" + originalName,
		ContentType:  contentType,
		Size:         size,
		Status:       domain.VideoStatusReady,
		OwnerID:      ownerID,
		Tags:         []string{},
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	f.videos[v.ID] = v
	return v, nil
}

func (f *fakeService) GetVideo(_ context.Context, id string) (domain.Video, error) {
	v, ok := f.videos[id]
	if !ok {
		return domain.Video{}, domain.ErrNotFound
	}
	return v, nil
}

func (f *fakeService) GetStreamURL(_ context.Context, id string) (string, error) {
	if _, ok := f.videos[id]; !ok {
		return "", domain.ErrNotFound
	}
	return "http://minio.local/stream/" + id, nil
}

func (f *fakeService) UpdateVideoStatus(_ context.Context, id string, status domain.VideoStatus) error {
	v, ok := f.videos[id]
	if !ok {
		return domain.ErrNotFound
	}
	v.Status = status
	v.UpdatedAt = time.Now().UTC()
	f.videos[id] = v
	return nil
}

func (f *fakeService) DeleteVideo(_ context.Context, id string) error {
	if _, ok := f.videos[id]; !ok {
		return domain.ErrNotFound
	}
	delete(f.videos, id)
	return nil
}

func (f *fakeService) ListOwnerVideos(_ context.Context, ownerID string) ([]domain.Video, error) {
	var result []domain.Video
	for _, v := range f.videos {
		if v.OwnerID == ownerID {
			result = append(result, v)
		}
	}
	if result == nil {
		result = []domain.Video{}
	}
	return result, nil
}

// ---- helpers ----------------------------------------------------------------

func buildUploadRequest(t *testing.T, ownerID, title, filename string, content []byte) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	_ = w.WriteField("owner_id", ownerID)
	_ = w.WriteField("title", title)
	fw, err := w.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	fw.Write(content)
	w.Close()
	req := httptest.NewRequest(http.MethodPost, "/videos", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	return req
}

func seedVideo(svc *fakeService, id, ownerID string) domain.Video {
	now := time.Now().UTC()
	v := domain.Video{
		ID:           id,
		Title:        "Test Video",
		OriginalName: "test.mp4",
		StoredKey:    id + "/test.mp4",
		ContentType:  "video/mp4",
		Size:         4096,
		Status:       domain.VideoStatusReady,
		OwnerID:      ownerID,
		Tags:         []string{},
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	svc.videos[id] = v
	return v
}

// ---- tests ------------------------------------------------------------------

// Test 1: GET /healthz returns 200 with {"status":"ok"}
func TestHealthz(t *testing.T) {
	svc := newFakeService()
	h := handler.New(svc)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body map[string]string
	json.NewDecoder(rec.Body).Decode(&body)
	if body["status"] != "ok" {
		t.Fatalf("expected status=ok, got %q", body["status"])
	}
}

// Test 2: POST /videos without owner_id returns 400
func TestUploadMissingOwnerID(t *testing.T) {
	svc := newFakeService()
	h := handler.New(svc)

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	_ = w.WriteField("title", "My Video")
	fw, _ := w.CreateFormFile("file", "v.mp4")
	fw.Write([]byte("data"))
	w.Close()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/videos", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

// Test 3: POST /videos with valid multipart returns 201 and video JSON
func TestUploadSuccess(t *testing.T) {
	svc := newFakeService()
	h := handler.New(svc)

	req := buildUploadRequest(t, "owner-1", "My Video", "clip.mp4", []byte("MP4 data"))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var video domain.Video
	if err := json.NewDecoder(rec.Body).Decode(&video); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if video.ID == "" {
		t.Fatal("expected non-empty ID")
	}
	if video.Title != "My Video" {
		t.Fatalf("expected 'My Video', got %q", video.Title)
	}
}

// Test 4: GET /videos/{id} for existing video returns 200 with metadata
func TestGetVideoFound(t *testing.T) {
	svc := newFakeService()
	seedVideo(svc, "vid-abc", "owner-1")
	h := handler.New(svc)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/videos/vid-abc", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var video domain.Video
	json.NewDecoder(rec.Body).Decode(&video)
	if video.ID != "vid-abc" {
		t.Fatalf("expected vid-abc, got %q", video.ID)
	}
}

// Test 5: GET /videos/{id} for missing video returns 404
func TestGetVideoNotFound(t *testing.T) {
	svc := newFakeService()
	h := handler.New(svc)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/videos/nope", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

// Test 6: GET /videos/{id}/stream returns 200 with streamUrl
func TestStreamVideo(t *testing.T) {
	svc := newFakeService()
	seedVideo(svc, "vid-stream", "owner-1")
	h := handler.New(svc)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/videos/vid-stream/stream", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body map[string]string
	json.NewDecoder(rec.Body).Decode(&body)
	if !strings.HasPrefix(body["streamUrl"], "http") {
		t.Fatalf("expected http URL, got %q", body["streamUrl"])
	}
}

// Test 7: PATCH /videos/{id}/status with valid status returns 204
func TestUpdateStatus(t *testing.T) {
	svc := newFakeService()
	seedVideo(svc, "vid-patch", "owner-1")
	h := handler.New(svc)

	body := `{"status":"PROCESSING"}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/videos/vid-patch/status", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

// Test 8: DELETE /videos/{id} returns 204; subsequent GET returns 404
func TestDeleteVideo(t *testing.T) {
	svc := newFakeService()
	seedVideo(svc, "vid-del", "owner-1")
	h := handler.New(svc)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/videos/vid-del", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}

	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/videos/vid-del", nil)
	h.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusNotFound {
		t.Fatalf("expected 404 after delete, got %d", rec2.Code)
	}
}
