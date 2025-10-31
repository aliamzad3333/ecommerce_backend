package models

import (
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ProductCategory represents the available product categories
type ProductCategory string

const (
	CategoryElectronics ProductCategory = "electronics"
	CategoryClothing    ProductCategory = "clothing"
	CategoryBooks       ProductCategory = "books"
	CategoryHome        ProductCategory = "home"
	CategorySports      ProductCategory = "sports"
	CategoryBeauty      ProductCategory = "beauty"
	CategoryToys        ProductCategory = "toys"
	CategoryAutomotive  ProductCategory = "automotive"
	CategoryFood        ProductCategory = "food"
	CategoryOther       ProductCategory = "other"
)

// GetValidCategories returns all valid product categories
func GetValidCategories() []ProductCategory {
	return []ProductCategory{
		CategoryElectronics,
		CategoryClothing,
		CategoryBooks,
		CategoryHome,
		CategorySports,
		CategoryBeauty,
		CategoryToys,
		CategoryAutomotive,
		CategoryFood,
		CategoryOther,
	}
}

// IsValidCategory checks if a category is valid
func IsValidCategory(category string) bool {
	for _, validCategory := range GetValidCategories() {
		if string(validCategory) == category {
			return true
		}
	}
	return false
}

// Product represents a product in the system
type Product struct {
	ID            primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name          string             `json:"name" bson:"name" validate:"required,min=2,max=100"`
	Price         float64            `json:"price" bson:"price" validate:"required,gt=0"`
	Category      ProductCategory    `json:"category" bson:"category" validate:"required"`
	ImageURL      string             `json:"image_url" bson:"image_url"`
	Description   string             `json:"description" bson:"description" validate:"required,min=10,max=1000"`
	Specification string             `json:"specification" bson:"specification"`
	Material      string             `json:"material" bson:"material"`
	InStock       bool               `json:"in_stock" bson:"in_stock"`
	CreatedBy     primitive.ObjectID `json:"created_by" bson:"created_by"`
	CreatedAt     time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at" bson:"updated_at"`
}

// CreateProductRequest represents the request payload for creating a product
type CreateProductRequest struct {
	Name          string  `json:"name" validate:"required,min=2,max=100"`
	Price         float64 `json:"price" validate:"required,gt=0"`
	Category      string  `json:"category" validate:"required"`
	Description   string  `json:"description" validate:"required,min=10,max=1000"`
	Specification string  `json:"specification"`
	Material      string  `json:"material"`
	InStock       bool    `json:"in_stock"`
}

// UpdateProductRequest represents the request payload for updating a product
type UpdateProductRequest struct {
	Name          *string  `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Price         *float64 `json:"price,omitempty" validate:"omitempty,gt=0"`
	Category      *string  `json:"category,omitempty"`
	Description   *string  `json:"description,omitempty" validate:"omitempty,min=10,max=1000"`
	Specification *string  `json:"specification,omitempty"`
	Material      *string  `json:"material,omitempty"`
	InStock       *bool    `json:"in_stock,omitempty"`
}

// ProductResponse represents the response payload for product operations
type ProductResponse struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Price         float64   `json:"price"`
	Category      string    `json:"category"`
	ImageURL      string    `json:"image_url"`
	Description   string    `json:"description"`
	Specification string    `json:"specification"`
	Material      string    `json:"material"`
	InStock       bool      `json:"in_stock"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ToResponse converts a Product to ProductResponse
func (p *Product) ToResponse() ProductResponse {
	return p.ToResponseWithBaseURL("")
}

// ToResponseWithBaseURL converts a Product to ProductResponse with a base URL for images
func (p *Product) ToResponseWithBaseURL(baseURL string) ProductResponse {
	imageURL := p.ImageURL
	// Convert relative URL to full URL if base URL is provided and image URL is relative
	if baseURL != "" && imageURL != "" && !strings.HasPrefix(imageURL, "http://") && !strings.HasPrefix(imageURL, "https://") {
		if strings.HasPrefix(imageURL, "/") {
			imageURL = baseURL + imageURL
		} else {
			imageURL = baseURL + "/" + imageURL
		}
	}

	return ProductResponse{
		ID:            p.ID.Hex(),
		Name:          p.Name,
		Price:         p.Price,
		Category:      string(p.Category),
		ImageURL:      imageURL,
		Description:   p.Description,
		Specification: p.Specification,
		Material:      p.Material,
		InStock:       p.InStock,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
	}
}

// ProductListResponse represents the response for listing products
type ProductListResponse struct {
	Products []ProductResponse `json:"products"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	Limit    int               `json:"limit"`
}

// CategoriesResponse represents the response for getting categories
type CategoriesResponse struct {
	Categories []string `json:"categories"`
}
