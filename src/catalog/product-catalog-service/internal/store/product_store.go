package store

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shopos/product-catalog-service/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const collectionName = "products"

// Collection is the subset of *mongo.Collection methods used by the store.
// It is defined as an interface to allow mocking in tests.
type Collection interface {
	InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error)
	FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) *mongo.SingleResult
	Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (cur *mongo.Cursor, err error)
	CountDocuments(ctx context.Context, filter interface{}, opts ...*options.CountOptions) (int64, error)
	FindOneAndUpdate(ctx context.Context, filter interface{}, update interface{}, opts ...*options.FindOneAndUpdateOptions) *mongo.SingleResult
}

// ProductStore handles all MongoDB operations for the products collection.
type ProductStore struct {
	col Collection
}

// New creates a ProductStore backed by the given MongoDB database.
func New(db *mongo.Database) *ProductStore {
	col := db.Collection(collectionName)
	return &ProductStore{col: col}
}

// NewWithCollection creates a ProductStore backed by the supplied Collection.
// Useful for unit tests that inject a mock.
func NewWithCollection(col Collection) *ProductStore {
	return &ProductStore{col: col}
}

// Create inserts a new product, generating a UUID for its ID.
func (s *ProductStore) Create(ctx context.Context, req *domain.CreateProductRequest) (*domain.Product, error) {
	now := time.Now().UTC()
	p := &domain.Product{
		ID:          uuid.New().String(),
		SKU:         req.SKU,
		Name:        req.Name,
		Description: req.Description,
		CategoryID:  req.CategoryID,
		BrandID:     req.BrandID,
		Price:       req.Price,
		Currency:    req.Currency,
		Status:      domain.StatusDraft,
		Tags:        req.Tags,
		Attributes:  req.Attributes,
		ImageURLs:   req.ImageURLs,
		Weight:      req.Weight,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if p.Tags == nil {
		p.Tags = []string{}
	}
	if p.ImageURLs == nil {
		p.ImageURLs = []string{}
	}
	if p.Attributes == nil {
		p.Attributes = map[string]string{}
	}

	_, err := s.col.InsertOne(ctx, p)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, domain.ErrDuplicateSKU
		}
		return nil, err
	}

	return p, nil
}

// GetByID retrieves a single product by its UUID string ID.
func (s *ProductStore) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	filter := bson.M{"_id": id, "status": bson.M{"$ne": string(domain.StatusArchived)}}
	return s.findOne(ctx, filter)
}

// GetBySKU retrieves a single product by its SKU.
func (s *ProductStore) GetBySKU(ctx context.Context, sku string) (*domain.Product, error) {
	filter := bson.M{"sku": sku, "status": bson.M{"$ne": string(domain.StatusArchived)}}
	return s.findOne(ctx, filter)
}

func (s *ProductStore) findOne(ctx context.Context, filter bson.M) (*domain.Product, error) {
	result := s.col.FindOne(ctx, filter)
	if err := result.Err(); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	var p domain.Product
	if err := result.Decode(&p); err != nil {
		return nil, err
	}

	return &p, nil
}

// List returns a filtered, paginated list of products.
func (s *ProductStore) List(ctx context.Context, req *domain.ListProductsRequest) (*domain.ProductList, error) {
	filter := bson.M{}

	if req.CategoryID != "" {
		filter["category_id"] = req.CategoryID
	}
	if req.BrandID != "" {
		filter["brand_id"] = req.BrandID
	}
	if req.Status != "" {
		filter["status"] = string(req.Status)
	} else {
		// By default exclude archived products from listings.
		filter["status"] = bson.M{"$ne": string(domain.StatusArchived)}
	}
	if req.MinPrice > 0 || req.MaxPrice > 0 {
		priceFilter := bson.M{}
		if req.MinPrice > 0 {
			priceFilter["$gte"] = req.MinPrice
		}
		if req.MaxPrice > 0 {
			priceFilter["$lte"] = req.MaxPrice
		}
		filter["price"] = priceFilter
	}
	if len(req.Tags) > 0 {
		filter["tags"] = bson.M{"$all": req.Tags}
	}

	total, err := s.col.CountDocuments(ctx, filter)
	if err != nil {
		return nil, err
	}

	limit := int64(req.Limit)
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	offset := int64(req.Offset)
	if offset < 0 {
		offset = 0
	}

	findOpts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(limit).
		SetSkip(offset)

	cursor, err := s.col.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var items []*domain.Product
	if err := cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	if items == nil {
		items = []*domain.Product{}
	}

	return &domain.ProductList{
		Items:  items,
		Total:  total,
		Limit:  int(limit),
		Offset: int(offset),
	}, nil
}

// Update applies a partial update to the product identified by id.
func (s *ProductStore) Update(ctx context.Context, id string, req *domain.UpdateProductRequest) (*domain.Product, error) {
	setFields := bson.M{
		"updated_at": time.Now().UTC(),
	}

	if req.Name != nil {
		setFields["name"] = *req.Name
	}
	if req.Description != nil {
		setFields["description"] = *req.Description
	}
	if req.Price != nil {
		setFields["price"] = *req.Price
	}
	if req.Status != nil {
		setFields["status"] = string(*req.Status)
	}
	if req.Tags != nil {
		setFields["tags"] = req.Tags
	}
	if req.Attributes != nil {
		setFields["attributes"] = req.Attributes
	}

	filter := bson.M{"_id": id}
	update := bson.M{"$set": setFields}

	after := options.After
	opt := options.FindOneAndUpdate().SetReturnDocument(after)

	result := s.col.FindOneAndUpdate(ctx, filter, update, opt)
	if err := result.Err(); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	var p domain.Product
	if err := result.Decode(&p); err != nil {
		return nil, err
	}

	return &p, nil
}

// Delete soft-deletes a product by setting its status to archived.
func (s *ProductStore) Delete(ctx context.Context, id string) error {
	archived := string(domain.StatusArchived)
	req := &domain.UpdateProductRequest{Status: (*domain.ProductStatus)(&archived)}
	_, err := s.Update(ctx, id, req)
	return err
}
