package domain

// Catalog
type Product struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Price       float64           `json:"price"`
	Currency    string            `json:"currency"`
	CategoryID  string            `json:"category_id"`
	Attributes  map[string]string `json:"attributes,omitempty"`
}

type ProductList struct {
	Items    []*Product `json:"items"`
	Total    int64      `json:"total"`
	Page     int        `json:"page"`
	PageSize int        `json:"page_size"`
}

type Category struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// Inventory
type StockLevel struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
	Reserved  int    `json:"reserved"`
	Available int    `json:"available"`
	Warehouse string `json:"warehouse,omitempty"`
}

// Orders
type OrderItem struct {
	ProductID string  `json:"product_id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
}

type Order struct {
	ID        string       `json:"id"`
	PartnerID string       `json:"partner_id"`
	Items     []*OrderItem `json:"items"`
	Total     float64      `json:"total"`
	Status    string       `json:"status"`
	CreatedAt string       `json:"created_at"`
}

type PlaceOrderRequest struct {
	Items     []OrderItem `json:"items"`
	AddressID string      `json:"address_id"`
}

// Webhooks
type Webhook struct {
	ID        string   `json:"id"`
	URL       string   `json:"url"`
	Events    []string `json:"events"`
	Active    bool     `json:"active"`
	CreatedAt string   `json:"created_at"`
}

type CreateWebhookRequest struct {
	URL    string   `json:"url"`
	Events []string `json:"events"`
}

// B2B
type Organization struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Country string `json:"country"`
	Status  string `json:"status"`
}

type Contract struct {
	ID         string  `json:"id"`
	PartnerID  string  `json:"partner_id"`
	Title      string  `json:"title"`
	Status     string  `json:"status"`
	ValidUntil string  `json:"valid_until"`
	Value      float64 `json:"value"`
}

type Quote struct {
	ID        string       `json:"id"`
	PartnerID string       `json:"partner_id"`
	Items     []*OrderItem `json:"items"`
	Status    string       `json:"status"`
	ExpiresAt string       `json:"expires_at"`
	Total     float64      `json:"total"`
}

type CreateQuoteRequest struct {
	Items []OrderItem `json:"items"`
	Note  string      `json:"note,omitempty"`
}
