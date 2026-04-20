package domain

// WarmingEvent is the decoded payload received from Kafka topics.
type WarmingEvent struct {
	ProductID  string `json:"product_id"`
	CartID     string `json:"cart_id"`
	OrderID    string `json:"order_id"`
	UserID     string `json:"user_id"`
	Query      string `json:"query"`
	StreamType string `json:"stream_type"`
}
