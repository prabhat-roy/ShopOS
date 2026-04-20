package handler_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/shopos/product-import-service/domain"
	"github.com/shopos/product-import-service/handler"
)

// ---- mock service -----------------------------------------------------------

type mockService struct {
	jobs       map[string]*domain.ImportJob
	createErr  error
	listErr    error
	processErr error
}

func newMockService() *mockService {
	return &mockService{jobs: make(map[string]*domain.ImportJob)}
}

func (m *mockService) CreateJob(fileName string, format domain.ImportFormat) (*domain.ImportJob, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	job := &domain.ImportJob{
		ID:        "test-job-id",
		FileName:  fileName,
		Format:    format,
		Status:    domain.ImportPending,
		Errors:    []domain.ImportError{},
		CreatedAt: time.Now().UTC(),
	}
	m.jobs[job.ID] = job
	return job, nil
}

func (m *mockService) GetJob(id string) (*domain.ImportJob, error) {
	job, ok := m.jobs[id]
	if !ok {
		return nil, fmt.Errorf("import job not found: %s", id)
	}
	return job, nil
}

func (m *mockService) ListJobs() ([]*domain.ImportJob, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	list := make([]*domain.ImportJob, 0, len(m.jobs))
	for _, j := range m.jobs {
		list = append(list, j)
	}
	return list, nil
}

func (m *mockService) ProcessCSV(jobID string, data []byte) error {
	return m.processErr
}

func (m *mockService) ProcessJSON(jobID string, data []byte) error {
	return m.processErr
}

// ---- helpers ----------------------------------------------------------------

func newHandler(svc handler.Servicer) *http.ServeMux {
	mux := http.NewServeMux()
	h := handler.New(svc)
	h.RegisterRoutes(mux)
	return mux
}

func doRequest(mux http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	var bodyBytes []byte
	if body != nil {
		bodyBytes, _ = json.Marshal(body)
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	return w
}

func decodeBody(w *httptest.ResponseRecorder) map[string]any {
	var m map[string]any
	_ = json.NewDecoder(w.Body).Decode(&m)
	return m
}

// ---- tests ------------------------------------------------------------------

func TestHealth(t *testing.T) {
	mux := newHandler(newMockService())
	w := doRequest(mux, http.MethodGet, "/healthz", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	body := decodeBody(w)
	if body["status"] != "ok" {
		t.Errorf("expected status=ok, got %v", body["status"])
	}
}

func TestCreateJob_Success(t *testing.T) {
	mux := newHandler(newMockService())
	payload := map[string]string{"file_name": "products.csv", "format": "CSV"}
	w := doRequest(mux, http.MethodPost, "/imports", payload)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d — body: %s", w.Code, w.Body.String())
	}
	body := decodeBody(w)
	if body["id"] == nil {
		t.Error("response should contain id field")
	}
	if body["status"] != string(domain.ImportPending) {
		t.Errorf("expected status=PENDING, got %v", body["status"])
	}
}

func TestCreateJob_MissingFileName(t *testing.T) {
	mux := newHandler(newMockService())
	payload := map[string]string{"format": "CSV"}
	w := doRequest(mux, http.MethodPost, "/imports", payload)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestCreateJob_InvalidFormat(t *testing.T) {
	mux := newHandler(newMockService())
	payload := map[string]string{"file_name": "products.xlsx", "format": "XLSX"}
	w := doRequest(mux, http.MethodPost, "/imports", payload)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestCreateJob_BadJSON(t *testing.T) {
	mux := newHandler(newMockService())
	req := httptest.NewRequest(http.MethodPost, "/imports", strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestListJobs(t *testing.T) {
	svc := newMockService()
	// Pre-seed two jobs.
	_, _ = svc.CreateJob("a.csv", domain.FormatCSV)
	_, _ = svc.CreateJob("b.json", domain.FormatJSON)
	mux := newHandler(svc)

	w := doRequest(mux, http.MethodGet, "/imports", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var jobs []map[string]any
	_ = json.NewDecoder(w.Body).Decode(&jobs)
	if len(jobs) != 2 {
		t.Errorf("expected 2 jobs, got %d", len(jobs))
	}
}

func TestListJobs_Empty(t *testing.T) {
	mux := newHandler(newMockService())
	w := doRequest(mux, http.MethodGet, "/imports", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var jobs []map[string]any
	_ = json.NewDecoder(w.Body).Decode(&jobs)
	if len(jobs) != 0 {
		t.Errorf("expected 0 jobs, got %d", len(jobs))
	}
}

func TestGetJob_Found(t *testing.T) {
	svc := newMockService()
	job, _ := svc.CreateJob("products.csv", domain.FormatCSV)
	mux := newHandler(svc)

	w := doRequest(mux, http.MethodGet, "/imports/"+job.ID, nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	body := decodeBody(w)
	if body["id"] != job.ID {
		t.Errorf("expected job id %s, got %v", job.ID, body["id"])
	}
}

func TestGetJob_NotFound(t *testing.T) {
	mux := newHandler(newMockService())
	w := doRequest(mux, http.MethodGet, "/imports/nonexistent-id", nil)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestProcessJob_CSV_Accepted(t *testing.T) {
	svc := newMockService()
	job, _ := svc.CreateJob("products.csv", domain.FormatCSV)
	mux := newHandler(svc)

	csvData := "sku,name,price\nSKU001,Widget,9.99\n"
	encoded := base64.StdEncoding.EncodeToString([]byte(csvData))
	payload := map[string]string{
		"format":      "CSV",
		"data_base64": encoded,
	}
	w := doRequest(mux, http.MethodPost, "/imports/"+job.ID+"/process", payload)

	if w.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d — body: %s", w.Code, w.Body.String())
	}
	body := decodeBody(w)
	if body["job_id"] != job.ID {
		t.Errorf("expected job_id=%s in response, got %v", job.ID, body["job_id"])
	}
}

func TestProcessJob_JSON_Accepted(t *testing.T) {
	svc := newMockService()
	job, _ := svc.CreateJob("products.json", domain.FormatJSON)
	mux := newHandler(svc)

	jsonData := `[{"sku":"SKU001","name":"Widget","price":9.99}]`
	encoded := base64.StdEncoding.EncodeToString([]byte(jsonData))
	payload := map[string]string{
		"format":      "JSON",
		"data_base64": encoded,
	}
	w := doRequest(mux, http.MethodPost, "/imports/"+job.ID+"/process", payload)

	if w.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestProcessJob_MissingData(t *testing.T) {
	svc := newMockService()
	job, _ := svc.CreateJob("products.csv", domain.FormatCSV)
	mux := newHandler(svc)

	payload := map[string]string{"format": "CSV"}
	w := doRequest(mux, http.MethodPost, "/imports/"+job.ID+"/process", payload)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestProcessJob_InvalidBase64(t *testing.T) {
	svc := newMockService()
	job, _ := svc.CreateJob("products.csv", domain.FormatCSV)
	mux := newHandler(svc)

	payload := map[string]string{"format": "CSV", "data_base64": "!!!not-base64!!!"}
	w := doRequest(mux, http.MethodPost, "/imports/"+job.ID+"/process", payload)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestProcessJob_JobNotFound(t *testing.T) {
	mux := newHandler(newMockService())

	encoded := base64.StdEncoding.EncodeToString([]byte("sku,name,price\n"))
	payload := map[string]string{"format": "CSV", "data_base64": encoded}
	w := doRequest(mux, http.MethodPost, "/imports/ghost-id/process", payload)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestMethodNotAllowed_Imports(t *testing.T) {
	mux := newHandler(newMockService())
	w := doRequest(mux, http.MethodDelete, "/imports", nil)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}
