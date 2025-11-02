package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// OrderStatus represents the status of an order
type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "pending"
	OrderStatusConfirmed  OrderStatus = "confirmed"
	OrderStatusProcessing OrderStatus = "processing"
	OrderStatusShipped    OrderStatus = "shipped"
	OrderStatusDelivered  OrderStatus = "delivered"
	OrderStatusCancelled  OrderStatus = "cancelled"
	OrderStatusRefunded   OrderStatus = "refunded"
)

// OrderItem represents an item in an order
type OrderItem struct {
	ProductID   primitive.ObjectID `json:"product_id" bson:"product_id"`
	ProductName string             `json:"product_name" bson:"product_name"`
	Quantity    int                `json:"quantity" bson:"quantity"`
	Price       float64            `json:"price" bson:"price"` // Price at time of order
	Subtotal    float64            `json:"subtotal" bson:"subtotal"`
}

// ShippingAddress represents the shipping address for an order
type ShippingAddress struct {
	FullName    string `json:"full_name" bson:"full_name"`
	AddressLine1 string `json:"address_line1" bson:"address_line1"`
	AddressLine2 string `json:"address_line2,omitempty" bson:"address_line2,omitempty"`
	City         string `json:"city" bson:"city"`
	State        string `json:"state" bson:"state"`
	PostalCode   string `json:"postal_code" bson:"postal_code"`
	Country      string `json:"country" bson:"country"`
	Phone        string `json:"phone" bson:"phone"`
}

// Order represents an order in the system
type Order struct {
	ID               primitive.ObjectID  `json:"-" bson:"_id,omitempty"` // MongoDB internal ID
	OrderID          string              `json:"order_id" bson:"order_id"` // Custom order ID (e.g., broshopbd_000001)
	UserID           *primitive.ObjectID `json:"user_id,omitempty" bson:"user_id,omitempty"` // Optional: null for guest orders
	GuestName        string              `json:"guest_name,omitempty" bson:"guest_name,omitempty"` // Name for guest orders (from shipping address)
	GuestEmail       string              `json:"guest_email,omitempty" bson:"guest_email,omitempty"` // Optional email for guest orders
	Items            []OrderItem         `json:"items" bson:"items"`
	ShippingAddress  ShippingAddress    `json:"shipping_address" bson:"shipping_address"`
	Subtotal         float64            `json:"subtotal" bson:"subtotal"`
	ShippingCost     float64            `json:"shipping_cost" bson:"shipping_cost"`
	Tax              float64            `json:"tax" bson:"tax"`
	Total            float64            `json:"total" bson:"total"`
	Status           OrderStatus        `json:"status" bson:"status"`
	PaymentMethod    string             `json:"payment_method" bson:"payment_method"`
	PaymentStatus    string             `json:"payment_status" bson:"payment_status"` // pending, paid, failed, refunded
	OrderNotes       string             `json:"order_notes,omitempty" bson:"order_notes,omitempty"`
	AdminNotes       string             `json:"admin_notes,omitempty" bson:"admin_notes,omitempty"` // Admin-only notes
	CreatedAt        time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at" bson:"updated_at"`
	ShippedAt        *time.Time         `json:"shipped_at,omitempty" bson:"shipped_at,omitempty"`
	DeliveredAt      *time.Time         `json:"delivered_at,omitempty" bson:"delivered_at,omitempty"`
}

// CreateOrderRequest represents the request payload for creating an order
type CreateOrderRequest struct {
	Items           []CreateOrderItemRequest `json:"items" validate:"required,min=1,dive"`
	ShippingAddress ShippingAddress          `json:"shipping_address" validate:"required"`
	PaymentMethod   string                   `json:"payment_method" validate:"required,oneof=cash card online"`
	OrderNotes      string                   `json:"order_notes,omitempty"`
	Subtotal        float64                  `json:"subtotal" validate:"gte=0"` // Frontend sends this (optional, defaults to 0)
	ShippingCost    float64                  `json:"shipping_cost" validate:"gte=0"` // Frontend sends this (optional, defaults to 0)
	Tax             float64                  `json:"tax" validate:"gte=0"` // Frontend sends this (optional, defaults to 0)
	Total           float64                  `json:"total" validate:"gte=0"` // Frontend sends this (optional, defaults to 0)
}

// CreateOrderItemRequest represents an item in the order creation request
type CreateOrderItemRequest struct {
	ProductID primitive.ObjectID `json:"product_id" validate:"required"`
	Quantity  int                `json:"quantity" validate:"required,min=1"`
}

// UpdateOrderRequest represents the request payload for updating an order
type UpdateOrderRequest struct {
	Status     *OrderStatus   `json:"status,omitempty" validate:"omitempty,oneof=pending confirmed processing shipped delivered cancelled refunded"`
	AdminNotes *string        `json:"admin_notes,omitempty"`
	ShippingAddress *ShippingAddress `json:"shipping_address,omitempty"` // User can update address if order is pending
}

// UpdateOrderStatusRequest represents the request to update order status
type UpdateOrderStatusRequest struct {
	Status     OrderStatus `json:"status" validate:"required,oneof=pending confirmed processing shipped delivered cancelled refunded"`
	AdminNotes *string     `json:"admin_notes,omitempty"`
}

// OrderResponse represents the response payload for order operations
type OrderResponse struct {
	OrderID          string          `json:"order_id"` // Custom order ID (e.g., broshopbd_000001)
	UserID           string          `json:"user_id,omitempty"` // Empty for guest orders
	GuestName        string          `json:"guest_name,omitempty"` // Present for guest orders (from shipping address)
	GuestEmail       string          `json:"guest_email,omitempty"` // Present for guest orders if provided
	Items            []OrderItem     `json:"items"`
	ShippingAddress  ShippingAddress `json:"shipping_address"`
	Subtotal         float64         `json:"subtotal"`
	ShippingCost     float64         `json:"shipping_cost"`
	Tax              float64         `json:"tax"`
	Total            float64         `json:"total"`
	Status           string          `json:"status"`
	PaymentMethod    string          `json:"payment_method"`
	PaymentStatus    string          `json:"payment_status"`
	OrderNotes       string          `json:"order_notes"` // Always included (empty string if not set)
	AdminNotes       string          `json:"admin_notes"` // Always included (empty string if not set)
	OrderPlacedDate  time.Time       `json:"order_placed_date"` // Date when order was placed (same as created_at)
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
	ShippedAt        *time.Time      `json:"shipped_at,omitempty"`
	DeliveredAt      *time.Time      `json:"delivered_at,omitempty"`
}

// OrderListResponse represents the response for listing orders
type OrderListResponse struct {
	Orders      []OrderResponse       `json:"orders"`
	Total       int                   `json:"total"`
	Page        int                   `json:"page"`
	Limit       int                   `json:"limit"`
	StatusCount map[string]int        `json:"status_count"` // Count of orders by status in current page
}

// GetValidOrderStatuses returns all valid order statuses
func GetValidOrderStatuses() []OrderStatus {
	return []OrderStatus{
		OrderStatusPending,
		OrderStatusConfirmed,
		OrderStatusProcessing,
		OrderStatusShipped,
		OrderStatusDelivered,
		OrderStatusCancelled,
		OrderStatusRefunded,
	}
}

// ToResponse converts an Order to OrderResponse
func (o *Order) ToResponse() OrderResponse {
	var shippedAt, deliveredAt *time.Time
	if o.ShippedAt != nil {
		shippedAt = o.ShippedAt
	}
	if o.DeliveredAt != nil {
		deliveredAt = o.DeliveredAt
	}

	var userIDStr string
	if o.UserID != nil {
		userIDStr = o.UserID.Hex()
	}
	
	return OrderResponse{
		OrderID:         o.OrderID,
		UserID:          userIDStr,
		GuestName:       o.GuestName,
		GuestEmail:      o.GuestEmail,
		Items:           o.Items,
		ShippingAddress: o.ShippingAddress,
		Subtotal:        o.Subtotal,
		ShippingCost:    o.ShippingCost,
		Tax:             o.Tax,
		Total:           o.Total,
		Status:          string(o.Status),
		PaymentMethod:   o.PaymentMethod,
		PaymentStatus:   o.PaymentStatus,
		OrderNotes:      o.OrderNotes,
		AdminNotes:      o.AdminNotes,
		OrderPlacedDate: o.CreatedAt, // Order placed date (same as created_at)
		CreatedAt:       o.CreatedAt,
		UpdatedAt:       o.UpdatedAt,
		ShippedAt:       shippedAt,
		DeliveredAt:     deliveredAt,
	}
}

