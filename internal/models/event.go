package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"gorm.io/gorm"
)

// Event represents an event in the system
type Event struct {
	ID               primitive.ObjectID `json:"id" bson:"_id,omitempty" gorm:"type:objectid;primaryKey;autoIncrement:false"`
	Title            string             `json:"title" bson:"title" gorm:"not null" validate:"required,min=3,max=200"`
	Description      string             `json:"description" bson:"description" gorm:"not null" validate:"required,max=1000"`
	Date             string             `json:"date" bson:"date" gorm:"not null" validate:"required"`
	Time             string             `json:"time,omitempty" bson:"time,omitempty"`
	Location         string             `json:"location" bson:"location" gorm:"not null" validate:"required,max=200"`
	Capacity         int                `json:"capacity" bson:"capacity" gorm:"not null" validate:"required,min=1"`
	Price            float64            `json:"price,omitempty" bson:"price,omitempty"`
	Category         string             `json:"category,omitempty" bson:"category,omitempty"`
	Organizer        string             `json:"organizer,omitempty" bson:"organizer,omitempty"`
	TicketsAvailable bool               `json:"tickets_available" bson:"tickets_available" gorm:"default:true"`
	Featured         bool               `json:"featured" bson:"featured" gorm:"default:false"`
	Published        bool               `json:"published" bson:"published" gorm:"default:false"`
	ImageURL         string             `json:"image_url,omitempty" bson:"image_url,omitempty"`
	CreatedAt        time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at" bson:"updated_at"`
}

// BeforeCreate hook to set ID and timestamps
func (e *Event) BeforeCreate(tx *gorm.DB) error {
	if e.ID.IsZero() {
		e.ID = primitive.NewObjectID()
	}
	e.CreatedAt = time.Now()
	e.UpdatedAt = time.Now()
	return nil
}

// BeforeUpdate hook to update timestamp
func (e *Event) BeforeUpdate(tx *gorm.DB) error {
	e.UpdatedAt = time.Now()
	return nil
}

// EventResponse represents event data returned to client
type EventResponse struct {
	ID               string    `json:"id"`
	Title            string    `json:"title"`
	Description      string    `json:"description"`
	Date             string    `json:"date"`
	Time             string    `json:"time"`
	Location         string    `json:"location"`
	Capacity         int       `json:"capacity"`
	Price            float64   `json:"price,omitempty"`
	Category         string    `json:"category,omitempty"`
	Organizer        string    `json:"organizer,omitempty"`
	TicketsAvailable bool      `json:"tickets_available"`
	Featured         bool      `json:"featured"`
	Published        bool      `json:"published"`
	ImageURL         string    `json:"image_url,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// ToResponse converts Event to EventResponse
func (e *Event) ToResponse() EventResponse {
	return EventResponse{
		ID:               e.ID.Hex(),
		Title:            e.Title,
		Description:      e.Description,
		Date:             e.Date,
		Time:             e.Time,
		Location:         e.Location,
		Capacity:         e.Capacity,
		Price:            e.Price,
		Category:         e.Category,
		Organizer:        e.Organizer,
		TicketsAvailable: e.TicketsAvailable,
		Featured:         e.Featured,
		Published:        e.Published,
		ImageURL:         e.ImageURL,
		CreatedAt:        e.CreatedAt,
		UpdatedAt:        e.UpdatedAt,
	}
}

// CreateEventRequest represents event creation request payload
type CreateEventRequest struct {
	Title            string  `json:"title" validate:"required,min=3,max=200"`
	Description      string  `json:"description" validate:"required,max=1000"`
	Date             string  `json:"date" validate:"required"`
	Time             string  `json:"time,omitempty"`
	Location         string  `json:"location" validate:"required,max=200"`
	Capacity         int     `json:"capacity" validate:"required,min=1"`
	Price            float64 `json:"price,omitempty"`
	Category         string  `json:"category,omitempty"`
	Organizer        string  `json:"organizer,omitempty"`
	TicketsAvailable bool    `json:"tickets_available"`
	Featured         bool    `json:"featured"`
	Published        bool    `json:"published"`
	ImageURL         string  `json:"image_url,omitempty"`
}

// UpdateEventRequest represents event update request payload
type UpdateEventRequest struct {
	Title            string  `json:"title,omitempty" validate:"omitempty,min=3,max=200"`
	Description      string  `json:"description,omitempty" validate:"omitempty,max=1000"`
	Date             string  `json:"date,omitempty"`
	Time             string  `json:"time,omitempty"`
	Location         string  `json:"location,omitempty" validate:"omitempty,max=200"`
	Capacity         int     `json:"capacity,omitempty" validate:"omitempty,min=1"`
	Price            float64 `json:"price,omitempty"`
	Category         string  `json:"category,omitempty"`
	Organizer        string  `json:"organizer,omitempty"`
	TicketsAvailable *bool   `json:"tickets_available,omitempty"`
	Featured         *bool   `json:"featured,omitempty"`
	Published        *bool   `json:"published,omitempty"`
	ImageURL         string  `json:"image_url,omitempty"`
}
