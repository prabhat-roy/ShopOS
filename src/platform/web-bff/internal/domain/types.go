package domain

// Auth
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	UserID       string `json:"user_id"`
}

type RegisterRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type RegisterResponse struct {
	UserID  string `json:"user_id"`
	Email   string `json:"email"`
	Message string `json:"message"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// Catalog
type Product struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Price       float64           `json:"price"`
	Currency    string            `json:"currency"`
	CategoryID  string            `json:"category_id"`
	BrandID     string            `json:"brand_id,omitempty"`
	Images      []string          `json:"images"`
	InStock     bool              `json:"in_stock"`
	Quantity    int               `json:"quantity"`
	Attributes  map[string]string `json:"attributes,omitempty"`
}

type ProductList struct {
	Items    []*Product `json:"items"`
	Total    int64      `json:"total"`
	Page     int        `json:"page"`
	PageSize int        `json:"page_size"`
}

type ListProductsRequest struct {
	Page       int     `json:"page"`
	PageSize   int     `json:"page_size"`
	CategoryID string  `json:"category_id,omitempty"`
	MinPrice   float64 `json:"min_price,omitempty"`
	MaxPrice   float64 `json:"max_price,omitempty"`
	Sort       string  `json:"sort,omitempty"`
}

type Category struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	ParentID string `json:"parent_id,omitempty"`
	Slug     string `json:"slug"`
}

// Cart
type CartItem struct {
	ItemID    string  `json:"item_id"`
	ProductID string  `json:"product_id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
	ImageURL  string  `json:"image_url,omitempty"`
}

type Cart struct {
	UserID    string      `json:"user_id"`
	Items     []*CartItem `json:"items"`
	Total     float64     `json:"total"`
	ItemCount int         `json:"item_count"`
}

type AddItemRequest struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type UpdateItemRequest struct {
	Quantity int `json:"quantity"`
}

// Order
type OrderItem struct {
	ProductID string  `json:"product_id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
}

type Order struct {
	ID          string       `json:"id"`
	UserID      string       `json:"user_id"`
	Items       []*OrderItem `json:"items"`
	Total       float64      `json:"total"`
	Status      string       `json:"status"`
	CreatedAt   string       `json:"created_at"`
	UpdatedAt   string       `json:"updated_at"`
}

type PlaceOrderRequest struct {
	AddressID string `json:"address_id"`
	PaymentID string `json:"payment_id"`
}

// User
type UserProfile struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

type UpdateProfileRequest struct {
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Phone     string `json:"phone,omitempty"`
}
