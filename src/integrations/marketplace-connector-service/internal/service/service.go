package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	kafkago "github.com/segmentio/kafka-go"
	"github.com/shopos/marketplace-connector-service/internal/adapter"
	"github.com/shopos/marketplace-connector-service/internal/domain"
	"github.com/shopos/marketplace-connector-service/internal/store"
)

// Servicer encapsulates marketplace sync business logic.
type Servicer struct {
	store   *store.SyncStore
	adapter *adapter.MarketplaceAdapter
	writer  *kafkago.Writer
	topic   string
}

// New constructs a Servicer.
func New(st *store.SyncStore, ad *adapter.MarketplaceAdapter, writer *kafkago.Writer, topic string) *Servicer {
	return &Servicer{store: st, adapter: ad, writer: writer, topic: topic}
}

// SyncProducts pushes the provided products to the given marketplace and records the result.
func (s *Servicer) SyncProducts(marketplace domain.Marketplace, products []map[string]interface{}) domain.SyncRecord {
	now := time.Now().UTC()
	rec := domain.SyncRecord{
		ID:          uuid.New().String(),
		Marketplace: marketplace,
		SyncType:    domain.SyncTypeProduct,
		Status:      domain.SyncStatusRunning,
		Errors:      []string{},
		StartedAt:   &now,
	}
	s.store.SaveSync(&rec)

	processed := 0
	failed := 0
	errs := []string{}

	for _, p := range products {
		var formatted map[string]interface{}
		switch marketplace {
		case domain.MarketplaceAmazon:
			formatted = s.adapter.FormatForAmazon(p)
		case domain.MarketplaceEbay:
			formatted = s.adapter.FormatForEbay(p)
		case domain.MarketplaceEtsy:
			formatted = s.adapter.FormatForEtsy(p)
		case domain.MarketplaceWalmart:
			formatted = s.adapter.FormatForWalmart(p)
		default:
			failed++
			errs = append(errs, fmt.Sprintf("unknown marketplace: %s", marketplace))
			continue
		}

		// Validate that the formatted payload has at minimum a title field.
		if formatted["Title"] == nil && formatted["title"] == nil && formatted["productName"] == nil {
			failed++
			errs = append(errs, fmt.Sprintf("product %v missing title", p["sku"]))
			continue
		}
		processed++
	}

	completedAt := time.Now().UTC()
	status := domain.SyncStatusCompleted
	if failed > 0 && processed == 0 {
		status = domain.SyncStatusFailed
	}

	rec.ItemsProcessed = processed
	rec.ItemsFailed = failed
	rec.Status = status
	rec.Errors = errs
	rec.CompletedAt = &completedAt
	s.store.SaveSync(&rec)

	s.publishEvent(rec)
	return rec
}

// SyncOrders simulates fetching orders from the marketplace and recording the sync.
func (s *Servicer) SyncOrders(marketplace domain.Marketplace, limit int) domain.SyncRecord {
	if limit <= 0 {
		limit = 50
	}

	now := time.Now().UTC()
	rec := domain.SyncRecord{
		ID:          uuid.New().String(),
		Marketplace: marketplace,
		SyncType:    domain.SyncTypeOrder,
		Status:      domain.SyncStatusRunning,
		Errors:      []string{},
		StartedAt:   &now,
	}
	s.store.SaveSync(&rec)

	// Simulate fetching orders: generate synthetic marketplace order payloads.
	fetched := simulateFetchOrders(marketplace, limit)
	processed := 0
	failed := 0
	errs := []string{}

	for _, raw := range fetched {
		order := s.adapter.ParseMarketplaceOrder(marketplace, raw)
		if order.MarketplaceOrderID == "" {
			failed++
			errs = append(errs, "order missing marketplace order id")
			continue
		}
		processed++
		_ = order // In a real system this would be forwarded to order-service via gRPC/Kafka.
	}

	completedAt := time.Now().UTC()
	status := domain.SyncStatusCompleted
	if failed > 0 && processed == 0 {
		status = domain.SyncStatusFailed
	}

	rec.ItemsProcessed = processed
	rec.ItemsFailed = failed
	rec.Status = status
	rec.Errors = errs
	rec.CompletedAt = &completedAt
	s.store.SaveSync(&rec)

	s.publishEvent(rec)
	return rec
}

// GetSync returns a sync record by ID.
func (s *Servicer) GetSync(id string) (*domain.SyncRecord, error) {
	return s.store.GetSync(id)
}

// ListSyncs returns sync records optionally filtered by marketplace.
func (s *Servicer) ListSyncs(marketplace domain.Marketplace, limit int) []*domain.SyncRecord {
	return s.store.ListSyncs(marketplace, limit)
}

// GetStats returns completed sync counts per marketplace.
func (s *Servicer) GetStats() map[domain.Marketplace]int {
	return s.store.GetStats()
}

// publishEvent sends a sync-completed event to Kafka.
func (s *Servicer) publishEvent(rec domain.SyncRecord) {
	payload, err := json.Marshal(rec)
	if err != nil {
		log.Printf("kafka marshal error: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = s.writer.WriteMessages(ctx, kafkago.Message{
		Key:   []byte(rec.ID),
		Value: payload,
		Headers: []kafkago.Header{
			{Key: "marketplace", Value: []byte(rec.Marketplace)},
			{Key: "syncType", Value: []byte(rec.SyncType)},
		},
	})
	if err != nil {
		log.Printf("kafka write error (non-fatal): %v", err)
	}
}

// simulateFetchOrders produces synthetic marketplace order maps for testing/simulation.
func simulateFetchOrders(marketplace domain.Marketplace, limit int) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, limit)
	for i := 0; i < limit; i++ {
		var raw map[string]interface{}
		switch marketplace {
		case domain.MarketplaceAmazon:
			raw = map[string]interface{}{
				"AmazonOrderId": fmt.Sprintf("112-%09d-%09d", i, i+1),
				"OrderStatus":   "Unshipped",
				"OrderTotal":    49.99 + float64(i),
				"OrderItems": []interface{}{
					map[string]interface{}{
						"OrderItemId":      fmt.Sprintf("item-%d", i),
						"SellerSKU":        fmt.Sprintf("SKU-%04d", i),
						"Title":            fmt.Sprintf("Product %d", i),
						"QuantityOrdered":  1,
						"ItemPrice":        49.99 + float64(i),
					},
				},
			}
		case domain.MarketplaceEbay:
			raw = map[string]interface{}{
				"OrderID":     fmt.Sprintf("ebay-order-%d", i),
				"OrderStatus": "Active",
				"Total":       29.99 + float64(i),
			}
		case domain.MarketplaceEtsy:
			raw = map[string]interface{}{
				"receipt_id": fmt.Sprintf("etsy-%d", i),
				"status":     "paid",
				"grandtotal": 19.99 + float64(i),
			}
		case domain.MarketplaceWalmart:
			raw = map[string]interface{}{
				"purchaseOrderId": fmt.Sprintf("WM-%d", i),
				"orderStatus":     "Acknowledged",
				"orderTotal":      39.99 + float64(i),
			}
		}
		result = append(result, raw)
	}
	return result
}
