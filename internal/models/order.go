package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"gorm.io/gorm"
)

type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "pending"
	OrderStatusConfirmed  OrderStatus = "confirmed"
	OrderStatusDelivered  OrderStatus = "delivered"
	OrderStatusCancelled  OrderStatus = "cancelled"
)

type PaymentStatus string

const (
	PaymentStatusPending PaymentStatus = "pending"
	PaymentStatusPaid    PaymentStatus = "paid"
	PaymentStatusFailed  PaymentStatus = "failed"
)

// OrderItem represents an item in an order
type OrderItem struct {
	ID       primitive.ObjectID `json:"id" bson:"_id,omitempty" gorm:"type:objectid;primaryKey;autoIncrement:false"`
	OrderID  primitive.ObjectID `json:"order_id" bson:"order_id" gorm:"type:objectid;index"`
	Name     string             `json:"name" bson:"name" gorm:"not null"`
	Quantity int                `json:"quantity" bson:"quantity" gorm:"not null" validate:"required,min=1"`
	Price    float64            `json:"price" bson:"price" gorm:"not null" validate:"required,min=0"`
}

// Order represents an order in the system
type Order struct {
	ID             primitive.ObjectID `json:"id" bson:"_id,omitempty" gorm:"type:objectid;primaryKey;autoIncrement:false"`
	OrderNumber    string             `json:"order_number" bson:"order_number" gorm:"uniqueIndex;not null"`
	UserID         primitive.ObjectID `json:"user_id,omitempty" bson:"user_id,omitempty" gorm:"type:objectid;index"`
	User           *User              `json:"user,omitempty" bson:"user,omitempty" gorm:"foreignKey:UserID"`
	CustomerName   string             `json:"customer_name" bson:"customer_name" gorm:"not null" validate:"required,min=2,max=100"`
	CustomerPhone  string             `json:"customer_phone" bson:"customer_phone" gorm:"not null" validate:"required"`
	CustomerEmail  string             `json:"customer_email,omitempty" bson:"customer_email,omitempty"`
	TotalAmount    float64            `json:"total_amount" bson:"total_amount" gorm:"not null" validate:"required,min=0"`
	Status         OrderStatus        `json:"status" bson:"status" gorm:"not null;default:pending" validate:"required,oneof=pending confirmed delivered cancelled"`
	PaymentStatus  PaymentStatus      `json:"payment_status" bson:"payment_status" gorm:"not null;default:pending" validate:"required,oneof=pending paid failed"`
	SpecialRequest string             `json:"special_request,omitempty" bson:"special_request,omitempty"`
	Items          []OrderItem        `json:"items" bson:"items" gorm:"foreignKey:OrderID"`
	CreatedAt      time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at" bson:"updated_at"`
}

// BeforeCreate hook to set ID, order number and timestamps
func (o *Order) BeforeCreate(tx *gorm.DB) error {
	if o.ID.IsZero() {
		o.ID = primitive.NewObjectID()
	}
	if o.OrderNumber == "" {
		o.OrderNumber = "ORD-" + o.ID.Hex()[:8]
	}
	o.CreatedAt = time.Now()
	o.UpdatedAt = time.Now()
	return nil
}

// BeforeUpdate hook to update timestamp
func (o *Order) BeforeUpdate(tx *gorm.DB) error {
	o.UpdatedAt = time.Now()
	return nil
}

// OrderResponse represents order data returned to client
type OrderResponse struct {
	ID             string        `json:"id"`
	OrderNumber    string        `json:"order_number"`
	UserID         string        `json:"user_id,omitempty"`
	User           *UserResponse `json:"user,omitempty"`
	CustomerName   string        `json:"customer_name"`
	CustomerPhone  string        `json:"customer_phone"`
	CustomerEmail  string        `json:"customer_email,omitempty"`
	TotalAmount    float64       `json:"total_amount"`
	Status         OrderStatus   `json:"status"`
	PaymentStatus  PaymentStatus `json:"payment_status"`
	SpecialRequest string        `json:"special_request,omitempty"`
	Items          []OrderItem   `json:"items"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
}

// ToResponse converts Order to OrderResponse
func (o *Order) ToResponse() OrderResponse {
	var userResponse *UserResponse
	if o.User != nil {
		userResp := o.User.ToResponse()
		userResponse = &userResp
	}

	return OrderResponse{
		ID:             o.ID.Hex(),
		OrderNumber:    o.OrderNumber,
		UserID:         o.UserID.Hex(),
		User:           userResponse,
		CustomerName:   o.CustomerName,
		CustomerPhone:  o.CustomerPhone,
		CustomerEmail:  o.CustomerEmail,
		TotalAmount:    o.TotalAmount,
		Status:         o.Status,
		PaymentStatus:  o.PaymentStatus,
		SpecialRequest: o.SpecialRequest,
		Items:          o.Items,
		CreatedAt:      o.CreatedAt,
		UpdatedAt:      o.UpdatedAt,
	}
}

// CreateOrderRequest represents order creation request payload
type CreateOrderRequest struct {
	UserID         string      `json:"user_id,omitempty"`
	CustomerName   string      `json:"customer_name" validate:"required,min=2,max=100"`
	CustomerPhone  string      `json:"customer_phone" validate:"required"`
	CustomerEmail  string      `json:"customer_email,omitempty" validate:"omitempty,email"`
	Status         OrderStatus `json:"status,omitempty" validate:"omitempty,oneof=pending confirmed delivered cancelled"`
	PaymentStatus  PaymentStatus `json:"payment_status,omitempty" validate:"omitempty,oneof=pending paid failed"`
	SpecialRequest string      `json:"special_request,omitempty"`
	Items          []OrderItemRequest `json:"items" validate:"required,min=1,dive"`
}

// OrderItemRequest represents order item in request
type OrderItemRequest struct {
	Name     string  `json:"name" validate:"required"`
	Quantity int     `json:"quantity" validate:"required,min=1"`
	Price    float64 `json:"price" validate:"required,min=0"`
}

// UpdateOrderRequest represents order update request payload
type UpdateOrderRequest struct {
	CustomerName   string       `json:"customer_name,omitempty" validate:"omitempty,min=2,max=100"`
	CustomerPhone  string       `json:"customer_phone,omitempty"`
	CustomerEmail  string       `json:"customer_email,omitempty" validate:"omitempty,email"`
	Status         OrderStatus  `json:"status,omitempty" validate:"omitempty,oneof=pending confirmed delivered cancelled"`
	PaymentStatus  PaymentStatus `json:"payment_status,omitempty" validate:"omitempty,oneof=pending paid failed"`
	SpecialRequest string       `json:"special_request,omitempty"`
}
