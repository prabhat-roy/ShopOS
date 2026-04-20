package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shopos/media-asset-service/internal/domain"
	"github.com/shopos/media-asset-service/internal/handler"
)

// ---- fake service -----------------------------------------------------------

type fakeService struct {
	assets map[string]domain.Asset
}

func newFakeService() *fakeService {
	return &fakeService{assets: make(map[string]domain.Asset)}
}

func (f *fakeService) UploadAsset(_ context.Context, ownerID, originalName, contentType string, _ io.Reader, size int64) (domain.Asset, error) {
	a := domain.Asset{
		ID:           "test-id-1",
		OriginalName: originalName,
		StoredName:   "test-id-1/" + originalName,
		ContentType:  contentType,
		Size:         size,
		AssetType:    domain.AssetTypeFromContentType(contentType),
		Bucket:       "media-assets",
		Tags:         []string{},
		OwnerID:      ownerID,
		UploadedAt:   time.Now().UTC(),
	}
	f.assets[a.ID] = a
	return a, nil
}

func (f *fakeService) GetAsset(_ context.Context, id string) (domain.Asset, error) {
	a, ok := f.assets[id]
	if !ok {
		return domain.Asset{}, domain.ErrNotFound
	}
	return a, nil
}

func (f *fakeService) GetDownloadURL(_ context.Context, id string, _ time.Duration) (string, error) {
	if _, ok := f.assets[id]; !ok {
		return "", domain.ErrNotFound
	}
	return "http://minio.local/presigned/" + id, nil
}

func (f *fakeService) DownloadAsset(_ context.Context, id string) (io.ReadCloser, domain.Asset, error) {
	a, ok := f.assets[id]
	if !ok {
		return nil, domain.Asset{}, domain.ErrNotFound
	}
	return io.NopCloser(bytes.NewReader([]byte("file-bytes"))), a, nil
}

func (f *fakeService) DeleteAsset(_ context.Context, id string) error {
	if _, ok := f.assets[id]; !ok {
		return domain.ErrNotFound
	}
	delete(f.assets, id)
	return nil
}

// ---- helpers ----------------------------------------------------------------

func buildUploadRequest(t *testing.T, ownerID, filename, contentType string, content []byte) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	_ = w.WriteField("owner_id", ownerID)
	fw, err := w.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	fw.Write(content)
	w.Close()

	req := httptest.NewRequest(http.MethodPost, "/assets", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	return req
}

func seedAsset(svc *fakeService, id string) domain.Asset {
	a := domain.Asset{
		ID:           id,
		OriginalName: "photo.jpg",
		StoredName:   id + "/photo.jpg",
		ContentType:  "image/jpeg",
		Size:         1024,
		AssetType:    domain.AssetTypeImage,
		Bucket:       "media-assets",
		Tags:         []string{},
		OwnerID:      "owner-1",
		UploadedAt:   time.Now().UTC(),
	}
	svc.assets[id] = a
	return a
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
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("expected status=ok, got %q", body["status"])
	}
}

// Test 2: POST /assets without owner_id returns 400
func TestUploadMissingOwnerID(t *testing.T) {
	svc := newFakeService()
	h := handler.New(svc)

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, _ := w.CreateFormFile("file", "test.jpg")
	fw.Write([]byte("data"))
	w.Close()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/assets", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

// Test 3: POST /assets with valid multipart returns 201 with asset JSON
func TestUploadSuccess(t *testing.T) {
	svc := newFakeService()
	h := handler.New(svc)

	req := buildUploadRequest(t, "owner-1", "image.png", "image/png", []byte("PNG data"))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var asset domain.Asset
	if err := json.NewDecoder(rec.Body).Decode(&asset); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if asset.ID == "" {
		t.Fatal("expected non-empty ID")
	}
	if asset.OwnerID != "owner-1" {
		t.Fatalf("expected owner-1, got %q", asset.OwnerID)
	}
	if asset.AssetType != domain.AssetTypeImage {
		t.Fatalf("expected IMAGE, got %q", asset.AssetType)
	}
}

// Test 4: GET /assets/{id} for existing asset returns 200 with metadata
func TestGetAssetFound(t *testing.T) {
	svc := newFakeService()
	seedAsset(svc, "asset-abc")
	h := handler.New(svc)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/assets/asset-abc", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var asset domain.Asset
	if err := json.NewDecoder(rec.Body).Decode(&asset); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if asset.ID != "asset-abc" {
		t.Fatalf("expected asset-abc, got %q", asset.ID)
	}
}

// Test 5: GET /assets/{id} for missing asset returns 404
func TestGetAssetNotFound(t *testing.T) {
	svc := newFakeService()
	h := handler.New(svc)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/assets/does-not-exist", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

// Test 6: GET /assets/{id}/download for existing asset returns 302 redirect
func TestDownloadRedirect(t *testing.T) {
	svc := newFakeService()
	seedAsset(svc, "asset-xyz")
	h := handler.New(svc)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/assets/asset-xyz/download", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("expected 302, got %d", rec.Code)
	}
	location := rec.Header().Get("Location")
	if location == "" {
		t.Fatal("expected Location header")
	}
}

// Test 7: GET /assets/{id}/download for missing asset returns 404
func TestDownloadNotFound(t *testing.T) {
	svc := newFakeService()
	h := handler.New(svc)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/assets/ghost/download", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

// Test 8: DELETE /assets/{id} returns 204; subsequent GET returns 404
func TestDeleteAsset(t *testing.T) {
	svc := newFakeService()
	seedAsset(svc, "asset-del")
	h := handler.New(svc)

	// delete
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/assets/asset-del", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}

	// confirm gone
	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/assets/asset-del", nil)
	h.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusNotFound {
		t.Fatalf("expected 404 after delete, got %d", rec2.Code)
	}
}
