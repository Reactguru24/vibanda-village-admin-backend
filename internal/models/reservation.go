package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"gorm.io/gorm"
)

type ReservationStatus string

const (
	ReservationStatusPending   ReservationStatus = "pending"
	ReservationStatusConfirmed ReservationStatus = "confirmed"
	ReservationStatusCancelled ReservationStatus = "cancelled"
)

// Reservation represents a reservation in the system
type Reservation struct {
	ID              primitive.ObjectID `json:"id" bson:"_id,omitempty" gorm:"type:objectid;primaryKey;autoIncrement:false"`
	UserID          primitive.ObjectID `json:"user_id,omitempty" bson:"user_id,omitempty" gorm:"type:objectid;index"`
	User            *User              `json:"user,omitempty" bson:"user,omitempty" gorm:"foreignKey:UserID"`
	CustomerName    string             `json:"customer_name" bson:"customer_name" gorm:"not null" validate:"required,min=2,max=100"`
	CustomerPhone   string             `json:"customer_phone" bson:"customer_phone" gorm:"not null" validate:"required"`
	CustomerEmail   string             `json:"customer_email" bson:"customer_email" gorm:"not null" validate:"required,email"`
	Date            string             `json:"date" bson:"date" gorm:"not null" validate:"required"`
	Time            string             `json:"time" bson:"time" gorm:"not null" validate:"required"`
	Guests          int                `json:"guests" bson:"guests" gorm:"not null" validate:"required,min=1,max=20"`
	SpecialRequests string             `json:"special_requests,omitempty" bson:"special_requests,omitempty"`
	Status          ReservationStatus  `json:"status" bson:"status" gorm:"not null;default:pending" validate:"required,oneof=pending confirmed cancelled"`
	CreatedAt       time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at" bson:"updated_at"`
}

// BeforeCreate hook to set ID and timestamps
func (r *Reservation) BeforeCreate(tx *gorm.DB) error {
	if r.ID.IsZero() {
		r.ID = primitive.NewObjectID()
	}
	r.CreatedAt = time.Now()
	r.UpdatedAt = time.Now()
	return nil
}

// BeforeUpdate hook to update timestamp
func (r *Reservation) BeforeUpdate(tx *gorm.DB) error {
	r.UpdatedAt = time.Now()
	return nil
}

// ReservationResponse represents reservation data returned to client
type ReservationResponse struct {
	ID              string            `json:"id"`
	UserID          string            `json:"user_id,omitempty"`
	User            *UserResponse     `json:"user,omitempty"`
	CustomerName    string            `json:"customer_name"`
	CustomerPhone   string            `json:"customer_phone"`
	CustomerEmail   string            `json:"customer_email"`
	Date            string            `json:"date"`
	Time            string            `json:"time"`
	Guests          int               `json:"guests"`
	SpecialRequests string            `json:"special_requests,omitempty"`
	Status          ReservationStatus `json:"status"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

// ToResponse converts Reservation to ReservationResponse
func (r *Reservation) ToResponse() ReservationResponse {
	var userResponse *UserResponse
	if r.User != nil {
		userResp := r.User.ToResponse()
		userResponse = &userResp
	}

	return ReservationResponse{
		ID:              r.ID.Hex(),
		UserID:          r.UserID.Hex(),
		User:            userResponse,
		CustomerName:    r.CustomerName,
		CustomerPhone:   r.CustomerPhone,
		CustomerEmail:   r.CustomerEmail,
		Date:            r.Date,
		Time:            r.Time,
		Guests:          r.Guests,
		SpecialRequests: r.SpecialRequests,
		Status:          r.Status,
		CreatedAt:       r.CreatedAt,
		UpdatedAt:       r.UpdatedAt,
	}
}

// CreateReservationRequest represents reservation creation request payload
type CreateReservationRequest struct {
	UserID          string            `json:"user_id,omitempty"`
	CustomerName    string            `json:"customer_name" validate:"required,min=2,max=100"`
	CustomerPhone   string            `json:"customer_phone" validate:"required"`
	CustomerEmail   string            `json:"customer_email" validate:"required,email"`
	Date            string            `json:"date" validate:"required"`
	Time            string            `json:"time" validate:"required"`
	Guests          int               `json:"guests" validate:"required,min=1,max=20"`
	SpecialRequests string            `json:"special_requests,omitempty"`
	Status          ReservationStatus `json:"status,omitempty" validate:"omitempty,oneof=pending confirmed cancelled"`
}

// UpdateReservationRequest represents reservation update request payload
type UpdateReservationRequest struct {
	CustomerName    string            `json:"customer_name,omitempty" validate:"omitempty,min=2,max=100"`
	CustomerPhone   string            `json:"customer_phone,omitempty"`
	CustomerEmail   string            `json:"customer_email,omitempty" validate:"omitempty,email"`
	Date            string            `json:"date,omitempty"`
	Time            string            `json:"time,omitempty"`
	Guests          int               `json:"guests,omitempty" validate:"omitempty,min=1,max=20"`
	SpecialRequests string            `json:"special_requests,omitempty"`
	Status          ReservationStatus `json:"status,omitempty" validate:"omitempty,oneof=pending confirmed cancelled"`
}
