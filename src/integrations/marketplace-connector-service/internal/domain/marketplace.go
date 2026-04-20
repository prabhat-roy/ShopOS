package domain

import "time"

// Marketplace represents a supported external marketplace.
type Marketplace string

const (
	MarketplaceAmazon  Marketplace = "AMAZON"
	MarketplaceEbay    Marketplace = "EBAY"
	MarketplaceEtsy    Marketplace = "ETSY"
	MarketplaceWalmart Marketplace = "WALMART"
)

// ValidMarketplace returns true when m is a recognised marketplace.
func ValidMarketplace(m Marketplace) bool {
	switch m {
	case MarketplaceAmazon, MarketplaceEbay, MarketplaceEtsy, MarketplaceWalmart:
		return true
	}
	return false
}

// SyncType represents what kind of entity is being synchronised.
type SyncType string

const (
	SyncTypeProduct   SyncType = "PRODUCT"
	SyncTypeOrder     SyncType = "ORDER"
	SyncTypeInventory SyncType = "INVENTORY"
)

// SyncStatus is the lifecycle state of a sync run.
type SyncStatus string

const (
	SyncStatusPending   SyncStatus = "PENDING"
	SyncStatusRunning   SyncStatus = "RUNNING"
	SyncStatusCompleted SyncStatus = "COMPLETED"
	SyncStatusFailed    SyncStatus = "FAILED"
)

// SyncRecord captures the result of a single sync run against a marketplace.
type SyncRecord struct {
	ID              string      `json:"id"`
	Marketplace     Marketplace `json:"marketplace"`
	SyncType        SyncType    `json:"syncType"`
	ItemsProcessed  int         `json:"itemsProcessed"`
	ItemsFailed     int         `json:"itemsFailed"`
	Status          SyncStatus  `json:"status"`
	Errors          []string    `json:"errors"`
	StartedAt       *time.Time  `json:"startedAt,omitempty"`
	CompletedAt     *time.Time  `json:"completedAt,omitempty"`
}

// ProductListing represents a ShopOS product as listed on a marketplace.
type ProductListing struct {
	MarketplaceID    string      `json:"marketplaceId"`
	ShopOsProductID  string      `json:"shopOsProductId"`
	SKU              string      `json:"sku"`
	Title            string      `json:"title"`
	Price            float64     `json:"price"`
	Currency         string      `json:"currency"`
	Status           string      `json:"status"`
	Marketplace      Marketplace `json:"marketplace"`
}

// OrderItem is a single line inside a MarketplaceOrder.
type OrderItem struct {
	LineItemID  string  `json:"lineItemId"`
	SKU         string  `json:"sku"`
	Title       string  `json:"title"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unitPrice"`
	Currency    string  `json:"currency"`
}

// MarketplaceOrder represents an order fetched from a marketplace.
type MarketplaceOrder struct {
	MarketplaceOrderID string      `json:"marketplaceOrderId"`
	ShopOsOrderID      string      `json:"shopOsOrderId"`
	Marketplace        Marketplace `json:"marketplace"`
	Status             string      `json:"status"`
	Items              []OrderItem `json:"items"`
	TotalAmount        float64     `json:"totalAmount"`
	CreatedAt          time.Time   `json:"createdAt"`
}
