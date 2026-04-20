package store_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shopos/product-catalog-service/internal/domain"
	"github.com/shopos/product-catalog-service/internal/store"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ---------------------------------------------------------------------------
// Mock Collection
// ---------------------------------------------------------------------------

type mockSingleResult struct {
	doc interface{}
	err error
}

func (m *mockSingleResult) Err() error { return m.err }
func (m *mockSingleResult) Decode(v interface{}) error {
	if m.err != nil {
		return m.err
	}
	// Use bson round-trip to copy fields properly.
	raw, err := bson.Marshal(m.doc)
	if err != nil {
		return err
	}
	return bson.Unmarshal(raw, v)
}

type mockCollection struct {
	insertErr    error
	findOneDoc   interface{}
	findOneErr   error
	findDocs     []*domain.Product
	findErr      error
	countResult  int64
	countErr     error
	updateDoc    interface{}
	updateErr    error
}

func (m *mockCollection) InsertOne(_ context.Context, _ interface{}, _ ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	if m.insertErr != nil {
		return nil, m.insertErr
	}
	return &mongo.InsertOneResult{}, nil
}

func (m *mockCollection) FindOne(_ context.Context, _ interface{}, _ ...*options.FindOneOptions) *mongo.SingleResult {
	// We cannot construct *mongo.SingleResult directly, so we wrap in a real
	// SingleResult obtained from an in-memory document. For simplicity in tests
	// that only care about ErrNoDocuments we return a real SingleResult from a
	// failed NewSingleResultFromDocument call with our err embedded.
	//
	// Because mongo.SingleResult is a concrete struct and cannot be directly
	// constructed in a test-friendly way, we return a wrapped result via the
	// interface type in the store. The store accepts store.Collection which
	// returns *mongo.SingleResult, so our mock must also return *mongo.SingleResult.
	//
	// We work around this by creating a real SingleResult via the exported
	// mongo.NewSingleResultFromDocument helper (available in v1.11+).
	if m.findOneErr != nil {
		return mongo.NewSingleResultFromDocument(nil, m.findOneErr, nil)
	}
	return mongo.NewSingleResultFromDocument(m.findOneDoc, nil, nil)
}

func (m *mockCollection) Find(_ context.Context, _ interface{}, _ ...*options.FindOptions) (*mongo.Cursor, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	// Build a cursor from the slice of documents.
	var docs []interface{}
	for _, p := range m.findDocs {
		raw, _ := bson.Marshal(p)
		docs = append(docs, raw)
	}
	cursor, err := mongo.NewCursorFromDocuments(docs, nil, nil)
	return cursor, err
}

func (m *mockCollection) CountDocuments(_ context.Context, _ interface{}, _ ...*options.CountOptions) (int64, error) {
	return m.countResult, m.countErr
}

func (m *mockCollection) FindOneAndUpdate(_ context.Context, _ interface{}, _ interface{}, _ ...*options.FindOneAndUpdateOptions) *mongo.SingleResult {
	if m.updateErr != nil {
		return mongo.NewSingleResultFromDocument(nil, m.updateErr, nil)
	}
	return mongo.NewSingleResultFromDocument(m.updateDoc, nil, nil)
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestCreate_GeneratesUUID(t *testing.T) {
	col := &mockCollection{}
	s := store.NewWithCollection(col)

	req := &domain.CreateProductRequest{
		SKU:      "TEST-SKU-001",
		Name:     "Test Product",
		Price:    9.99,
		Currency: "USD",
	}

	p, err := s.Create(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if p.ID == "" {
		t.Fatal("expected a non-empty UUID id")
	}
	if len(p.ID) != 36 {
		t.Fatalf("expected UUID length 36, got %d (value: %s)", len(p.ID), p.ID)
	}
	if p.SKU != req.SKU {
		t.Errorf("expected SKU %q, got %q", req.SKU, p.SKU)
	}
	if p.Status != domain.StatusDraft {
		t.Errorf("expected initial status %q, got %q", domain.StatusDraft, p.Status)
	}
	if p.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
	if p.UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be set")
	}
}

func TestCreate_DuplicateSKU(t *testing.T) {
	// mongo.IsDuplicateKeyError checks for a WriteException with code 11000.
	dupErr := mongo.WriteException{
		WriteErrors: []mongo.WriteError{
			{Code: 11000, Message: "duplicate key error"},
		},
	}

	col := &mockCollection{insertErr: dupErr}
	s := store.NewWithCollection(col)

	_, err := s.Create(context.Background(), &domain.CreateProductRequest{SKU: "DUP"})
	if !errors.Is(err, domain.ErrDuplicateSKU) {
		t.Errorf("expected ErrDuplicateSKU, got %v", err)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	col := &mockCollection{findOneErr: mongo.ErrNoDocuments}
	s := store.NewWithCollection(col)

	_, err := s.GetByID(context.Background(), "nonexistent-id")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestGetByID_Success(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Millisecond)
	product := &domain.Product{
		ID:        "test-uuid-1234",
		SKU:       "SKU-001",
		Name:      "Widget",
		Status:    domain.StatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}

	col := &mockCollection{findOneDoc: product}
	s := store.NewWithCollection(col)

	got, err := s.GetByID(context.Background(), "test-uuid-1234")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != product.ID {
		t.Errorf("expected ID %q, got %q", product.ID, got.ID)
	}
	if got.Name != product.Name {
		t.Errorf("expected Name %q, got %q", product.Name, got.Name)
	}
}

func TestGetBySKU_NotFound(t *testing.T) {
	col := &mockCollection{findOneErr: mongo.ErrNoDocuments}
	s := store.NewWithCollection(col)

	_, err := s.GetBySKU(context.Background(), "UNKNOWN-SKU")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestList_ReturnsPagedResults(t *testing.T) {
	products := []*domain.Product{
		{ID: "1", SKU: "A", Name: "Alpha", Status: domain.StatusActive},
		{ID: "2", SKU: "B", Name: "Beta", Status: domain.StatusActive},
	}

	col := &mockCollection{
		findDocs:    products,
		countResult: 2,
	}
	s := store.NewWithCollection(col)

	list, err := s.List(context.Background(), &domain.ListProductsRequest{Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if list.Total != 2 {
		t.Errorf("expected total 2, got %d", list.Total)
	}
	if len(list.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(list.Items))
	}
}

func TestDelete_SoftDeletes(t *testing.T) {
	archived := domain.StatusArchived
	archivedProduct := &domain.Product{
		ID:     "del-id",
		Status: archived,
	}

	col := &mockCollection{updateDoc: archivedProduct}
	s := store.NewWithCollection(col)

	err := s.Delete(context.Background(), "del-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdate_NotFound(t *testing.T) {
	col := &mockCollection{updateErr: mongo.ErrNoDocuments}
	s := store.NewWithCollection(col)

	name := "New Name"
	_, err := s.Update(context.Background(), "missing-id", &domain.UpdateProductRequest{Name: &name})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
