package handler_test

import (
	"context"

	"github.com/shopos/partner-bff/internal/domain"
	"github.com/shopos/partner-bff/internal/service"
)

// mockCatalogService implements service.CatalogService
type mockCatalogService struct {
	products   *domain.ProductList
	product    *domain.Product
	categories []*domain.Category
	err        error
}

func (m *mockCatalogService) ListProducts(_ context.Context, _, _ int, _ string) (*domain.ProductList, error) {
	return m.products, m.err
}
func (m *mockCatalogService) GetProduct(_ context.Context, _ string) (*domain.Product, error) {
	return m.product, m.err
}
func (m *mockCatalogService) ListCategories(_ context.Context) ([]*domain.Category, error) {
	return m.categories, m.err
}

var _ service.CatalogService = (*mockCatalogService)(nil)

// mockInventoryService implements service.InventoryService
type mockInventoryService struct {
	stock     *domain.StockLevel
	bulkStock []*domain.StockLevel
	err       error
}

func (m *mockInventoryService) GetStock(_ context.Context, _ string) (*domain.StockLevel, error) {
	return m.stock, m.err
}
func (m *mockInventoryService) GetBulkStock(_ context.Context, _ []string) ([]*domain.StockLevel, error) {
	return m.bulkStock, m.err
}

var _ service.InventoryService = (*mockInventoryService)(nil)

// mockOrderService implements service.OrderService
type mockOrderService struct {
	orders []*domain.Order
	order  *domain.Order
	err    error
}

func (m *mockOrderService) ListOrders(_ context.Context, _ string, _, _ int) ([]*domain.Order, error) {
	return m.orders, m.err
}
func (m *mockOrderService) GetOrder(_ context.Context, _, _ string) (*domain.Order, error) {
	return m.order, m.err
}
func (m *mockOrderService) PlaceOrder(_ context.Context, _ string, _ *domain.PlaceOrderRequest) (*domain.Order, error) {
	return m.order, m.err
}

var _ service.OrderService = (*mockOrderService)(nil)

// mockWebhookService implements service.WebhookService
type mockWebhookService struct {
	webhooks []*domain.Webhook
	webhook  *domain.Webhook
	err      error
}

func (m *mockWebhookService) ListWebhooks(_ context.Context, _ string) ([]*domain.Webhook, error) {
	return m.webhooks, m.err
}
func (m *mockWebhookService) CreateWebhook(_ context.Context, _ string, _ *domain.CreateWebhookRequest) (*domain.Webhook, error) {
	return m.webhook, m.err
}
func (m *mockWebhookService) DeleteWebhook(_ context.Context, _, _ string) error {
	return m.err
}

var _ service.WebhookService = (*mockWebhookService)(nil)

// mockB2BService implements service.B2BService
type mockB2BService struct {
	org       *domain.Organization
	contracts []*domain.Contract
	quotes    []*domain.Quote
	quote     *domain.Quote
	err       error
}

func (m *mockB2BService) GetOrganization(_ context.Context, _ string) (*domain.Organization, error) {
	return m.org, m.err
}
func (m *mockB2BService) ListContracts(_ context.Context, _ string) ([]*domain.Contract, error) {
	return m.contracts, m.err
}
func (m *mockB2BService) ListQuotes(_ context.Context, _ string) ([]*domain.Quote, error) {
	return m.quotes, m.err
}
func (m *mockB2BService) CreateQuote(_ context.Context, _ string, _ *domain.CreateQuoteRequest) (*domain.Quote, error) {
	return m.quote, m.err
}

var _ service.B2BService = (*mockB2BService)(nil)
