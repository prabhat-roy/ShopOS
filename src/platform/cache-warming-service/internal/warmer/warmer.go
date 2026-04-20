package warmer

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/shopos/cache-warming-service/internal/cache"
	"github.com/shopos/cache-warming-service/internal/domain"
)

// Warmer pre-populates Redis keys based on observed Kafka events.
type Warmer struct {
	cache cache.Cacher
	ttl   time.Duration
}

func New(c cache.Cacher, ttl time.Duration) *Warmer {
	return &Warmer{cache: c, ttl: ttl}
}

// HandleProductViewed increments a popularity counter for the product and marks
// it as a warming candidate so downstream fetches find a warm key.
func (w *Warmer) HandleProductViewed(ctx context.Context, msg []byte) error {
	var ev domain.WarmingEvent
	if err := json.Unmarshal(msg, &ev); err != nil {
		return fmt.Errorf("product_viewed: %w", err)
	}
	if ev.ProductID == "" {
		return nil
	}

	popularKey := fmt.Sprintf("warm:product:popular:%s", ev.ProductID)
	n, err := w.cache.Incr(ctx, popularKey)
	if err != nil {
		return err
	}
	if n == 1 {
		_ = w.cache.Expire(ctx, popularKey, 24*time.Hour)
	}

	warmKey := fmt.Sprintf("warm:product:pending:%s", ev.ProductID)
	if err := w.cache.Set(ctx, warmKey, "1", w.ttl); err != nil {
		return err
	}

	slog.Info("product viewed", "product_id", ev.ProductID, "views", n)
	return nil
}

// HandleCartAbandoned marks the user's cart for warming so recovery flows are fast.
func (w *Warmer) HandleCartAbandoned(ctx context.Context, msg []byte) error {
	var ev domain.WarmingEvent
	if err := json.Unmarshal(msg, &ev); err != nil {
		return fmt.Errorf("cart_abandoned: %w", err)
	}
	if ev.CartID == "" {
		return nil
	}

	key := fmt.Sprintf("warm:cart:abandoned:%s", ev.CartID)
	if err := w.cache.Set(ctx, key, "1", w.ttl); err != nil {
		return err
	}

	slog.Info("cart abandoned", "cart_id", ev.CartID)
	return nil
}

// HandleOrderPlaced warms the order lookup key so the confirmation page is instant.
func (w *Warmer) HandleOrderPlaced(ctx context.Context, msg []byte) error {
	var ev domain.WarmingEvent
	if err := json.Unmarshal(msg, &ev); err != nil {
		return fmt.Errorf("order_placed: %w", err)
	}
	if ev.OrderID == "" {
		return nil
	}

	key := fmt.Sprintf("warm:order:%s", ev.OrderID)
	if err := w.cache.Set(ctx, key, "1", w.ttl); err != nil {
		return err
	}

	slog.Info("order placed warm", "order_id", ev.OrderID)
	return nil
}

// HandleInventoryLow marks the product's inventory key for refresh so stock
// counts stay accurate without a cache stampede.
func (w *Warmer) HandleInventoryLow(ctx context.Context, msg []byte) error {
	var ev domain.WarmingEvent
	if err := json.Unmarshal(msg, &ev); err != nil {
		return fmt.Errorf("inventory_low: %w", err)
	}
	if ev.ProductID == "" {
		return nil
	}

	key := fmt.Sprintf("warm:inventory:low:%s", ev.ProductID)
	if err := w.cache.Set(ctx, key, "1", w.ttl); err != nil {
		return err
	}

	slog.Info("inventory low warm", "product_id", ev.ProductID)
	return nil
}

// HandleSearchPerformed increments a popularity counter for the search query
// so the search-service can pre-warm its result cache for hot queries.
func (w *Warmer) HandleSearchPerformed(ctx context.Context, msg []byte) error {
	var ev domain.WarmingEvent
	if err := json.Unmarshal(msg, &ev); err != nil {
		return fmt.Errorf("search_performed: %w", err)
	}
	if ev.Query == "" {
		return nil
	}

	popularKey := fmt.Sprintf("warm:search:popular:%s", ev.Query)
	n, err := w.cache.Incr(ctx, popularKey)
	if err != nil {
		return err
	}
	if n == 1 {
		_ = w.cache.Expire(ctx, popularKey, 24*time.Hour)
	}

	slog.Info("search performed", "query", ev.Query, "count", n)
	return nil
}
