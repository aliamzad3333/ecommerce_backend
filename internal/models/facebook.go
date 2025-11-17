package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// FacebookToken represents a Facebook Conversions API access token
type FacebookToken struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	AccessToken string             `bson:"access_token" json:"access_token" validate:"required"`
	TokenType   string             `bson:"token_type,omitempty" json:"token_type,omitempty"` // e.g., "Bearer"
	IsActive    bool               `bson:"is_active" json:"is_active"`
	CreatedBy   string             `bson:"created_by" json:"created_by"` // User email or ID
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

// FacebookTokenRequest represents the request payload for saving/updating a token
type FacebookTokenRequest struct {
	AccessToken string `json:"access_token" validate:"required"`
	TokenType   string `json:"token_type,omitempty"`
}

// FacebookTokenResponse represents the response payload for token operations
type FacebookTokenResponse struct {
	ID          primitive.ObjectID `json:"id"`
	IsActive    bool               `json:"is_active"`
	CreatedBy   string             `json:"created_by"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
	// Note: AccessToken is NOT included in response for security
}

// ToResponse converts FacebookToken to FacebookTokenResponse (without exposing the token)
func (ft *FacebookToken) ToResponse() FacebookTokenResponse {
	return FacebookTokenResponse{
		ID:        ft.ID,
		IsActive:  ft.IsActive,
		CreatedBy: ft.CreatedBy,
		CreatedAt: ft.CreatedAt,
		UpdatedAt: ft.UpdatedAt,
	}
}

