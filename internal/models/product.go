package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"gorm.io/gorm"
)

type ProductCategory string

const (
	CategoryFood  ProductCategory = "food"
	CategoryDrink ProductCategory = "drink"
)

type ProductSubcategory string

// Food subcategories
const (
	SubcategoryMain     ProductSubcategory = "main"
	SubcategoryStarters ProductSubcategory = "starters"
	SubcategoryDessert  ProductSubcategory = "dessert"
)

// Drink subcategories
const (
	SubcategoryBeer  ProductSubcategory = "beer"
	SubcategoryWine  ProductSubcategory = "wine"
	SubcategoryJuice ProductSubcategory = "juice"
	SubcategoryOther ProductSubcategory = "other"
)

// Product represents a product in the system
type Product struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty" gorm:"type:objectid;primaryKey;autoIncrement:false"`
	Name         string             `json:"name" bson:"name" gorm:"not null" validate:"required,min=2,max=100"`
	Category     ProductCategory    `json:"category" bson:"category" gorm:"not null" validate:"required,oneof=food drink"`
	Subcategory  ProductSubcategory `json:"subcategory" bson:"subcategory" gorm:"not null" validate:"required"`
	Price        float64            `json:"price" bson:"price" gorm:"not null" validate:"required,min=0"`
	Stock        int                `json:"stock" bson:"stock" gorm:"not null;default:0" validate:"min=0"`
	Description  string             `json:"description,omitempty" bson:"description,omitempty" validate:"max=500"`
	ImageURL     string             `json:"image_url,omitempty" bson:"image_url,omitempty"`
	Popular      bool               `json:"popular" bson:"popular" gorm:"default:false"`
	New          bool               `json:"new" bson:"new" gorm:"default:false"`
	Available    bool               `json:"available" bson:"available" gorm:"default:true"`
	CreatedAt    time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at" bson:"updated_at"`
}

// BeforeCreate hook to set ID and timestamps
func (p *Product) BeforeCreate(tx *gorm.DB) error {
	if p.ID.IsZero() {
		p.ID = primitive.NewObjectID()
	}
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	return nil
}

// BeforeUpdate hook to update timestamp
func (p *Product) BeforeUpdate(tx *gorm.DB) error {
	p.UpdatedAt = time.Now()
	return nil
}

// ProductResponse represents product data returned to client
type ProductResponse struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Category    ProductCategory    `json:"category"`
	Subcategory ProductSubcategory `json:"subcategory"`
	Price       float64            `json:"price"`
	Stock       int                `json:"stock"`
	Description string             `json:"description,omitempty"`
	ImageURL    string             `json:"image_url,omitempty"`
	Popular     bool               `json:"popular"`
	New         bool               `json:"new"`
	Available   bool               `json:"available"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

// ToResponse converts Product to ProductResponse
func (p *Product) ToResponse() ProductResponse {
	return ProductResponse{
		ID:          p.ID.Hex(),
		Name:        p.Name,
		Category:    p.Category,
		Subcategory: p.Subcategory,
		Price:       p.Price,
		Stock:       p.Stock,
		Description: p.Description,
		ImageURL:    p.ImageURL,
		Popular:     p.Popular,
		New:         p.New,
		Available:   p.Available,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

// CreateProductRequest represents product creation request payload
type CreateProductRequest struct {
	Name        string             `json:"name" validate:"required,min=2,max=100"`
	Category    ProductCategory    `json:"category" validate:"required,oneof=food drink"`
	Subcategory ProductSubcategory `json:"subcategory" validate:"required"`
	Price       float64            `json:"price" validate:"required,min=0"`
	Stock       int                `json:"stock" validate:"min=0"`
	Description string             `json:"description,omitempty" validate:"max=500"`
	ImageURL    string             `json:"image_url,omitempty"`
	Popular     bool               `json:"popular"`
	New         bool               `json:"new"`
	Available   bool               `json:"available"`
}

// UpdateProductRequest represents product update request payload
type UpdateProductRequest struct {
	Name        string             `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Category    ProductCategory    `json:"category,omitempty" validate:"omitempty,oneof=food drink"`
	Subcategory ProductSubcategory `json:"subcategory,omitempty"`
	Price       float64            `json:"price,omitempty" validate:"omitempty,min=0"`
	Stock       int                `json:"stock,omitempty" validate:"min=0"`
	Description string             `json:"description,omitempty" validate:"max=500"`
	ImageURL    string             `json:"image_url,omitempty"`
	Popular     *bool              `json:"popular,omitempty"`
	New         *bool              `json:"new,omitempty"`
	Available   *bool              `json:"available,omitempty"`
}
