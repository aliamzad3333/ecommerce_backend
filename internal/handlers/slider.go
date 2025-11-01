package handlers

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ecommerce-backend/internal/database"
	"ecommerce-backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// SliderHandler handles slider-related HTTP requests
type SliderHandler struct {
	db        *database.Client
	validator *validator.Validate
}

// NewSliderHandler creates a new SliderHandler
func NewSliderHandler(db *database.Client) *SliderHandler {
	return &SliderHandler{
		db:        db,
		validator: validator.New(),
	}
}

// UploadSliderImage handles slider image upload (Admin only)
func (h *SliderHandler) UploadSliderImage(c *gin.Context) {
	userRole, exists := c.Get("user_role")
	if !exists || userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	// Parse multipart form
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No image file provided"})
		return
	}
	defer file.Close()

	// Validate file type
	if !isValidImageType(header.Filename) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid image format. Only JPG, JPEG, PNG, and GIF are allowed"})
		return
	}

	// Validate file size (max 5MB)
	if header.Size > 5*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Image size too large. Maximum 5MB allowed"})
		return
	}

	// Create uploads directory if it doesn't exist
	uploadDir := "uploads/slider"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
		return
	}

	// Generate unique filename
	filename := fmt.Sprintf("slider_%d%s", time.Now().Unix(), filepath.Ext(header.Filename))
	filePath := filepath.Join(uploadDir, filename)

	// Save file
	if err := saveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image"})
		return
	}

	// Get next order number (count existing sliders)
	collection := h.db.GetCollection("slider")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	count, err := collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		count = 0
	}

	// Save to database
	userID, _ := c.Get("user_id")
	adminID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		os.Remove(filePath)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	imageURL := fmt.Sprintf("/uploads/slider/%s", filename)
	slider := models.Slider{
		ID:        primitive.NewObjectID(),
		ImageURL:  imageURL,
		Order:     int(count), // Auto-increment order
		CreatedBy: adminID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err = collection.InsertOne(ctx, slider)
	if err != nil {
		os.Remove(filePath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save slider to database"})
		return
	}

	// Return full URL
	fullImageURL := slider.ImageURL
	baseURL := getBaseURL(c)
	if baseURL != "" && fullImageURL != "" && !strings.HasPrefix(fullImageURL, "http://") && !strings.HasPrefix(fullImageURL, "https://") {
		if strings.HasPrefix(fullImageURL, "/") {
			fullImageURL = baseURL + fullImageURL
		} else {
			fullImageURL = baseURL + "/" + fullImageURL
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Image uploaded successfully",
		"slider":  slider.ToResponseWithBaseURL(baseURL),
	})
}

// GetSliders retrieves all sliders for public display with settings
func (h *SliderHandler) GetSliders(c *gin.Context) {
	collection := h.db.GetCollection("slider")
	settingsCollection := h.db.GetCollection("slider_settings")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get all sliders, sorted by order
	findOptions := options.Find().SetSort(bson.D{{Key: "order", Value: 1}})
	cursor, err := collection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch sliders"})
		return
	}
	defer cursor.Close(ctx)

	var sliders []models.Slider
	if err = cursor.All(ctx, &sliders); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode sliders"})
		return
	}

	// Get slider settings
	var settings models.SliderSettings
	err = settingsCollection.FindOne(ctx, bson.M{}).Decode(&settings)
	if err == mongo.ErrNoDocuments {
		// Create default settings if none exist
		settings = models.SliderSettings{
			ID:             primitive.NewObjectID(),
			SlideDuration:  5, // Default 5 seconds
			AutoPlay:       true,
			ShowIndicators: true,
			ShowControls:   true,
			UpdatedAt:      time.Now(),
		}
		settingsCollection.InsertOne(ctx, settings)
	}

	// Convert to response format
	baseURL := getBaseURL(c)
	var sliderResponses []models.SliderResponse
	for _, slider := range sliders {
		sliderResponses = append(sliderResponses, slider.ToResponseWithBaseURL(baseURL))
	}

	response := models.PublicSliderResponse{
		Slides: sliderResponses,
		Settings: models.SliderSettingsResponse{
			SlideDuration:  settings.SlideDuration,
			AutoPlay:       settings.AutoPlay,
			ShowIndicators: settings.ShowIndicators,
			ShowControls:   settings.ShowControls,
			UpdatedAt:      settings.UpdatedAt,
		},
		TotalSlides: len(sliderResponses),
	}

	c.JSON(http.StatusOK, response)
}

// GetAllSliders retrieves all sliders for admin (list view)
func (h *SliderHandler) GetAllSliders(c *gin.Context) {
	userRole, exists := c.Get("user_role")
	if !exists || userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	collection := h.db.GetCollection("slider")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	findOptions := options.Find().SetSort(bson.D{{Key: "order", Value: 1}})
	cursor, err := collection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch sliders"})
		return
	}
	defer cursor.Close(ctx)

	var sliders []models.Slider
	if err = cursor.All(ctx, &sliders); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode sliders"})
		return
	}

	baseURL := getBaseURL(c)
	var sliderResponses []models.SliderResponse
	for _, slider := range sliders {
		sliderResponses = append(sliderResponses, slider.ToResponseWithBaseURL(baseURL))
	}

	c.JSON(http.StatusOK, gin.H{
		"sliders":      sliderResponses,
		"total_slides": len(sliderResponses),
	})
}

// DeleteSlider deletes a slider image (Admin only)
func (h *SliderHandler) DeleteSlider(c *gin.Context) {
	userRole, exists := c.Get("user_role")
	if !exists || userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	sliderID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(sliderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid slider ID"})
		return
	}

	collection := h.db.GetCollection("slider")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get slider to delete image file
	var slider models.Slider
	err = collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&slider)
	if err == nil && slider.ImageURL != "" {
		// Delete image file
		filename := filepath.Base(slider.ImageURL)
		imagePath := filepath.Join("uploads", "slider", filename)
		os.Remove(imagePath) // Ignore error if file doesn't exist
	}

	// Delete from database
	result, err := collection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete slider"})
		return
	}

	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Slider not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Slider deleted successfully"})
}

// GetSliderSettings retrieves slider settings (Admin only)
func (h *SliderHandler) GetSliderSettings(c *gin.Context) {
	userRole, exists := c.Get("user_role")
	if !exists || userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	collection := h.db.GetCollection("slider_settings")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var settings models.SliderSettings
	err := collection.FindOne(ctx, bson.M{}).Decode(&settings)
	if err == mongo.ErrNoDocuments {
		// Return default settings if none exist
		c.JSON(http.StatusOK, models.SliderSettingsResponse{
			SlideDuration:  5,
			AutoPlay:      true,
			ShowIndicators: true,
			ShowControls:   true,
			UpdatedAt:     time.Now(),
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch settings"})
		return
	}

	c.JSON(http.StatusOK, models.SliderSettingsResponse{
		SlideDuration:  settings.SlideDuration,
		AutoPlay:       settings.AutoPlay,
		ShowIndicators: settings.ShowIndicators,
		ShowControls:   settings.ShowControls,
		UpdatedAt:      settings.UpdatedAt,
	})
}

// UpdateSliderSettings updates slider settings (Admin only)
func (h *SliderHandler) UpdateSliderSettings(c *gin.Context) {
	userRole, exists := c.Get("user_role")
	if !exists || userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	var req models.UpdateSliderSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	collection := h.db.GetCollection("slider_settings")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update := bson.M{"$set": bson.M{"updated_at": time.Now()}}

	if req.SlideDuration != nil {
		update["$set"].(bson.M)["slide_duration"] = *req.SlideDuration
	}
	if req.AutoPlay != nil {
		update["$set"].(bson.M)["auto_play"] = *req.AutoPlay
	}
	if req.ShowIndicators != nil {
		update["$set"].(bson.M)["show_indicators"] = *req.ShowIndicators
	}
	if req.ShowControls != nil {
		update["$set"].(bson.M)["show_controls"] = *req.ShowControls
	}

	// Upsert settings (create if doesn't exist)
	opts := options.Update().SetUpsert(true)
	_, err := collection.UpdateOne(ctx, bson.M{}, update, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update settings"})
		return
	}

	// Fetch updated settings
	var settings models.SliderSettings
	err = collection.FindOne(ctx, bson.M{}).Decode(&settings)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch updated settings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Settings updated successfully",
		"settings": models.SliderSettingsResponse{
			SlideDuration:  settings.SlideDuration,
			AutoPlay:       settings.AutoPlay,
			ShowIndicators: settings.ShowIndicators,
			ShowControls:   settings.ShowControls,
			UpdatedAt:      settings.UpdatedAt,
		},
	})
}

