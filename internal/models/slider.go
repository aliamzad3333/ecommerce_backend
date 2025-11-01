package models

import (
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Slider represents a slider slide (just image)
type Slider struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ImageURL  string             `json:"image_url" bson:"image_url"`
	Order     int                `json:"order" bson:"order"` // Display order
	CreatedBy primitive.ObjectID `json:"created_by" bson:"created_by"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
}

// SliderResponse represents the response payload for slider operations
type SliderResponse struct {
	ID        string    `json:"id"`
	ImageURL  string    `json:"image_url"`
	Order     int       `json:"order"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SliderSettings represents slider settings (duration, etc.)
type SliderSettings struct {
	ID             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	SlideDuration  int                `json:"slide_duration" bson:"slide_duration"` // Duration in seconds
	AutoPlay       bool               `json:"auto_play" bson:"auto_play"`
	ShowIndicators bool               `json:"show_indicators" bson:"show_indicators"`
	ShowControls   bool               `json:"show_controls" bson:"show_controls"`
	UpdatedAt      time.Time          `json:"updated_at" bson:"updated_at"`
}

// UpdateSliderSettingsRequest represents the request to update slider settings
type UpdateSliderSettingsRequest struct {
	SlideDuration  *int  `json:"slide_duration,omitempty" validate:"omitempty,min=1,max=30"`
	AutoPlay       *bool `json:"auto_play,omitempty"`
	ShowIndicators *bool `json:"show_indicators,omitempty"`
	ShowControls   *bool `json:"show_controls,omitempty"`
}

// SliderSettingsResponse represents slider settings response
type SliderSettingsResponse struct {
	SlideDuration  int       `json:"slide_duration"`
	AutoPlay       bool      `json:"auto_play"`
	ShowIndicators bool      `json:"show_indicators"`
	ShowControls   bool      `json:"show_controls"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// PublicSliderResponse represents the public API response with slides and settings
type PublicSliderResponse struct {
	Slides      []SliderResponse       `json:"slides"`
	Settings    SliderSettingsResponse `json:"settings"`
	TotalSlides int                    `json:"total_slides"`
}

// ToResponse converts a Slider to SliderResponse
func (s *Slider) ToResponse() SliderResponse {
	return SliderResponse{
		ID:        s.ID.Hex(),
		ImageURL:  s.ImageURL,
		Order:     s.Order,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}
}

// ToResponseWithBaseURL converts a Slider to SliderResponse with full image URL
func (s *Slider) ToResponseWithBaseURL(baseURL string) SliderResponse {
	imageURL := s.ImageURL
	// Convert relative URL to full URL if base URL is provided
	if baseURL != "" && imageURL != "" && !strings.HasPrefix(imageURL, "http://") && !strings.HasPrefix(imageURL, "https://") {
		if strings.HasPrefix(imageURL, "/") {
			imageURL = baseURL + imageURL
		} else {
			imageURL = baseURL + "/" + imageURL
		}
	}

	return SliderResponse{
		ID:        s.ID.Hex(),
		ImageURL:  imageURL,
		Order:     s.Order,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}
}
