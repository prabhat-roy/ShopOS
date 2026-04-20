package handler_test

import (
	"context"

	"github.com/shopos/web-bff/internal/domain"
)

// --- Auth mock ---
type mockAuthService struct {
	loginFn   func(ctx context.Context, email, password string) (*domain.LoginResponse, error)
	registerFn func(ctx context.Context, req *domain.RegisterRequest) (*domain.RegisterResponse, error)
	refreshFn  func(ctx context.Context, token string) (*domain.TokenPair, error)
}

func (m *mockAuthService) Login(ctx context.Context, email, password string) (*domain.LoginResponse, error) {
	return m.loginFn(ctx, email, password)
}
func (m *mockAuthService) Register(ctx context.Context, req *domain.RegisterRequest) (*domain.RegisterResponse, error) {
	return m.registerFn(ctx, req)
}
func (m *mockAuthService) RefreshToken(ctx context.Context, token string) (*domain.TokenPair, error) {
	return m.refreshFn(ctx, token)
}

// --- Catalog mock ---
type mockCatalogService struct {
	listFn       func(ctx context.Context, req *domain.ListProductsRequest) (*domain.ProductList, error)
	getFn        func(ctx context.Context, id string) (*domain.Product, error)
	categoriesFn func(ctx context.Context) ([]*domain.Category, error)
	searchFn     func(ctx context.Context, q string, page, pageSize int) (*domain.ProductList, error)
}

func (m *mockCatalogService) ListProducts(ctx context.Context, req *domain.ListProductsRequest) (*domain.ProductList, error) {
	return m.listFn(ctx, req)
}
func (m *mockCatalogService) GetProduct(ctx context.Context, id string) (*domain.Product, error) {
	return m.getFn(ctx, id)
}
func (m *mockCatalogService) ListCategories(ctx context.Context) ([]*domain.Category, error) {
	return m.categoriesFn(ctx)
}
func (m *mockCatalogService) Search(ctx context.Context, q string, page, pageSize int) (*domain.ProductList, error) {
	return m.searchFn(ctx, q, page, pageSize)
}

// --- Cart mock ---
type mockCartService struct {
	getFn    func(ctx context.Context, userID string) (*domain.Cart, error)
	addFn    func(ctx context.Context, userID, productID string, qty int) (*domain.Cart, error)
	updateFn func(ctx context.Context, userID, itemID string, qty int) (*domain.Cart, error)
	removeFn func(ctx context.Context, userID, itemID string) (*domain.Cart, error)
	clearFn  func(ctx context.Context, userID string) error
}

func (m *mockCartService) GetCart(ctx context.Context, userID string) (*domain.Cart, error) {
	return m.getFn(ctx, userID)
}
func (m *mockCartService) AddItem(ctx context.Context, userID, productID string, qty int) (*domain.Cart, error) {
	return m.addFn(ctx, userID, productID, qty)
}
func (m *mockCartService) UpdateItem(ctx context.Context, userID, itemID string, qty int) (*domain.Cart, error) {
	return m.updateFn(ctx, userID, itemID, qty)
}
func (m *mockCartService) RemoveItem(ctx context.Context, userID, itemID string) (*domain.Cart, error) {
	return m.removeFn(ctx, userID, itemID)
}
func (m *mockCartService) ClearCart(ctx context.Context, userID string) error {
	return m.clearFn(ctx, userID)
}
