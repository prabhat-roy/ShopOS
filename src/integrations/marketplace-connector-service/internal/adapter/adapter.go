package adapter

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopos/marketplace-connector-service/internal/domain"
)

// MarketplaceAdapter translates between ShopOS internal structures and
// marketplace-specific field layouts.
type MarketplaceAdapter struct{}

// New returns a new MarketplaceAdapter.
func New() *MarketplaceAdapter {
	return &MarketplaceAdapter{}
}

// FormatForAmazon converts a generic product map to Amazon listing fields.
func (a *MarketplaceAdapter) FormatForAmazon(product map[string]interface{}) map[string]interface{} {
	price, _ := toFloat64(product["price"])
	quantity, _ := toInt(product["quantity"])
	return map[string]interface{}{
		"ASIN":              product["marketplaceId"],
		"Title":             product["title"],
		"ListPrice":         price,
		"Quantity":          quantity,
		"SKU":               product["sku"],
		"ProductCondition":  "New",
		"FulfillmentChannel": "MFN",
		"BrowseNodeId":      product["categoryId"],
	}
}

// FormatForEbay converts a generic product map to eBay listing fields.
func (a *MarketplaceAdapter) FormatForEbay(product map[string]interface{}) map[string]interface{} {
	price, _ := toFloat64(product["price"])
	quantity, _ := toInt(product["quantity"])
	return map[string]interface{}{
		"ItemID":          product["marketplaceId"],
		"Title":           product["title"],
		"StartPrice":      price,
		"Quantity":        quantity,
		"SKU":             product["sku"],
		"ListingType":     "FixedPriceItem",
		"ListingDuration": "GTC",
		"ConditionID":     1000,
		"PrimaryCategory": map[string]interface{}{"CategoryID": product["categoryId"]},
	}
}

// FormatForEtsy converts a generic product map to Etsy listing fields.
func (a *MarketplaceAdapter) FormatForEtsy(product map[string]interface{}) map[string]interface{} {
	price, _ := toFloat64(product["price"])
	quantity, _ := toInt(product["quantity"])
	return map[string]interface{}{
		"listing_id":       product["marketplaceId"],
		"title":            product["title"],
		"price":            price,
		"quantity":         quantity,
		"sku":              []string{fmt.Sprintf("%v", product["sku"])},
		"state":            "active",
		"shipping_profile_id": product["shippingProfileId"],
		"taxonomy_id":      product["categoryId"],
		"who_made":         "i_did",
		"is_supply":        false,
		"when_made":        "made_to_order",
	}
}

// FormatForWalmart converts a generic product map to Walmart listing fields.
func (a *MarketplaceAdapter) FormatForWalmart(product map[string]interface{}) map[string]interface{} {
	price, _ := toFloat64(product["price"])
	quantity, _ := toInt(product["quantity"])
	return map[string]interface{}{
		"sku":           product["sku"],
		"productName":   product["title"],
		"price": map[string]interface{}{
			"currency":  product["currency"],
			"amount":    price,
		},
		"fulfillmentLagTime": 2,
		"productType":        product["categoryId"],
		"wpid":               product["marketplaceId"],
		"publishedStatus":    "PUBLISHED",
		"lifecycleStatus":    "ACTIVE",
		"availabilityCode":   "AC",
		"quantity": map[string]interface{}{
			"unit":   "EACH",
			"amount": quantity,
		},
	}
}

// ParseMarketplaceOrder converts a raw marketplace payload into a MarketplaceOrder.
func (a *MarketplaceAdapter) ParseMarketplaceOrder(marketplace domain.Marketplace, data map[string]interface{}) domain.MarketplaceOrder {
	order := domain.MarketplaceOrder{
		ShopOsOrderID: uuid.New().String(),
		Marketplace:   marketplace,
		CreatedAt:     time.Now().UTC(),
		Items:         []domain.OrderItem{},
	}

	switch marketplace {
	case domain.MarketplaceAmazon:
		order.MarketplaceOrderID = stringVal(data, "AmazonOrderId")
		order.Status = stringVal(data, "OrderStatus")
		order.TotalAmount = floatVal(data, "OrderTotal")
		if items, ok := data["OrderItems"].([]interface{}); ok {
			for _, raw := range items {
				if m, ok := raw.(map[string]interface{}); ok {
					order.Items = append(order.Items, domain.OrderItem{
						LineItemID: stringVal(m, "OrderItemId"),
						SKU:        stringVal(m, "SellerSKU"),
						Title:      stringVal(m, "Title"),
						Quantity:   intVal(m, "QuantityOrdered"),
						UnitPrice:  floatVal(m, "ItemPrice"),
						Currency:   "USD",
					})
				}
			}
		}

	case domain.MarketplaceEbay:
		order.MarketplaceOrderID = stringVal(data, "OrderID")
		order.Status = stringVal(data, "OrderStatus")
		order.TotalAmount = floatVal(data, "Total")
		if items, ok := data["TransactionArray"].([]interface{}); ok {
			for _, raw := range items {
				if m, ok := raw.(map[string]interface{}); ok {
					order.Items = append(order.Items, domain.OrderItem{
						LineItemID: stringVal(m, "TransactionID"),
						SKU:        stringVal(m, "SKU"),
						Title:      stringVal(m, "Item.Title"),
						Quantity:   intVal(m, "QuantityPurchased"),
						UnitPrice:  floatVal(m, "TransactionPrice"),
						Currency:   "USD",
					})
				}
			}
		}

	case domain.MarketplaceEtsy:
		order.MarketplaceOrderID = stringVal(data, "receipt_id")
		order.Status = stringVal(data, "status")
		order.TotalAmount = floatVal(data, "grandtotal")
		if items, ok := data["transactions"].([]interface{}); ok {
			for _, raw := range items {
				if m, ok := raw.(map[string]interface{}); ok {
					order.Items = append(order.Items, domain.OrderItem{
						LineItemID: stringVal(m, "transaction_id"),
						SKU:        stringVal(m, "sku"),
						Title:      stringVal(m, "title"),
						Quantity:   intVal(m, "quantity"),
						UnitPrice:  floatVal(m, "price"),
						Currency:   "USD",
					})
				}
			}
		}

	case domain.MarketplaceWalmart:
		order.MarketplaceOrderID = stringVal(data, "purchaseOrderId")
		order.Status = stringVal(data, "orderStatus")
		order.TotalAmount = floatVal(data, "orderTotal")
		if items, ok := data["orderLines"].([]interface{}); ok {
			for _, raw := range items {
				if m, ok := raw.(map[string]interface{}); ok {
					order.Items = append(order.Items, domain.OrderItem{
						LineItemID: stringVal(m, "lineNumber"),
						SKU:        stringVal(m, "item.sku"),
						Title:      stringVal(m, "item.itemName"),
						Quantity:   intVal(m, "orderLineQuantity.amount"),
						UnitPrice:  floatVal(m, "charges.chargeAmount.amount"),
						Currency:   "USD",
					})
				}
			}
		}
	}

	return order
}

// GetFieldMapping returns a canonical ShopOS field → marketplace field mapping.
func (a *MarketplaceAdapter) GetFieldMapping(marketplace domain.Marketplace) map[string]string {
	switch marketplace {
	case domain.MarketplaceAmazon:
		return map[string]string{
			"id":       "ASIN",
			"title":    "Title",
			"price":    "ListPrice",
			"sku":      "SKU",
			"quantity": "Quantity",
			"orderId":  "AmazonOrderId",
			"status":   "OrderStatus",
		}
	case domain.MarketplaceEbay:
		return map[string]string{
			"id":       "ItemID",
			"title":    "Title",
			"price":    "StartPrice",
			"sku":      "SKU",
			"quantity": "Quantity",
			"orderId":  "OrderID",
			"status":   "OrderStatus",
		}
	case domain.MarketplaceEtsy:
		return map[string]string{
			"id":       "listing_id",
			"title":    "title",
			"price":    "price",
			"sku":      "sku",
			"quantity": "quantity",
			"orderId":  "receipt_id",
			"status":   "status",
		}
	case domain.MarketplaceWalmart:
		return map[string]string{
			"id":       "wpid",
			"title":    "productName",
			"price":    "price.amount",
			"sku":      "sku",
			"quantity": "quantity.amount",
			"orderId":  "purchaseOrderId",
			"status":   "orderStatus",
		}
	}
	return map[string]string{}
}

// --- helpers ---

func toFloat64(v interface{}) (float64, bool) {
	if v == nil {
		return 0, false
	}
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	}
	return 0, false
}

func toInt(v interface{}) (int, bool) {
	if v == nil {
		return 0, false
	}
	switch n := v.(type) {
	case int:
		return n, true
	case float64:
		return int(n), true
	case int64:
		return int(n), true
	}
	return 0, false
}

func stringVal(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		return fmt.Sprintf("%v", v)
	}
	return ""
}

func floatVal(m map[string]interface{}, key string) float64 {
	v, _ := toFloat64(m[key])
	return v
}

func intVal(m map[string]interface{}, key string) int {
	v, _ := toInt(m[key])
	return v
}
