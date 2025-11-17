package handlers

import (
	"context"
	"net/http"
	"time"

	"ecommerce-backend/internal/database"
	"ecommerce-backend/internal/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// FacebookHandler handles Facebook Conversions API token operations
type FacebookHandler struct {
	db         *database.Client
	collection *mongo.Collection
}

// NewFacebookHandler creates a new Facebook handler
func NewFacebookHandler(db *database.Client) *FacebookHandler {
	return &FacebookHandler{
		db:         db,
		collection: db.GetCollection("facebook_tokens"),
	}
}

// SaveToken saves or updates the Facebook access token
// POST /api/admin/facebook/token
func (h *FacebookHandler) SaveToken(c *gin.Context) {
	// Get user from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req models.FacebookTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Validate token is not empty
	if req.AccessToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Access token is required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if token already exists
	var existingToken models.FacebookToken
	err := h.collection.FindOne(ctx, bson.M{"is_active": true}).Decode(&existingToken)

	now := time.Now()
	userIDStr := userID.(string)

	if err == nil {
		// Token exists, update it
		update := bson.M{
			"$set": bson.M{
				"access_token": req.AccessToken,
				"token_type":   req.TokenType,
				"updated_at":   now,
				"created_by":   userIDStr,
			},
		}

		_, err = h.collection.UpdateOne(ctx, bson.M{"_id": existingToken.ID}, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update token: " + err.Error()})
			return
		}

		// Get updated token
		err = h.collection.FindOne(ctx, bson.M{"_id": existingToken.ID}).Decode(&existingToken)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve updated token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Token updated successfully",
			"token":   existingToken.ToResponse(),
		})
		return
	}

	// No existing token, create new one
	// First, deactivate any existing tokens
	_, _ = h.collection.UpdateMany(ctx, bson.M{}, bson.M{"$set": bson.M{"is_active": false}})

	// Create new token
	newToken := models.FacebookToken{
		ID:          primitive.NewObjectID(),
		AccessToken: req.AccessToken,
		TokenType:   req.TokenType,
		IsActive:    true,
		CreatedBy:   userIDStr,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	_, err = h.collection.InsertOne(ctx, newToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save token: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Token saved successfully",
		"token":   newToken.ToResponse(),
	})
}

// GetToken retrieves the active Facebook access token (without exposing the actual token)
// GET /api/admin/facebook/token
func (h *FacebookHandler) GetToken(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var token models.FacebookToken
	err := h.collection.FindOne(ctx, bson.M{"is_active": true}).Decode(&token)

	if err == mongo.ErrNoDocuments {
		c.JSON(http.StatusOK, gin.H{
			"message": "No active token found",
			"token":   nil,
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve token: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token.ToResponse(),
	})
}

// GetTokenValue retrieves the actual token value (for internal use only)
// GET /api/admin/facebook/token/value
func (h *FacebookHandler) GetTokenValue(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var token models.FacebookToken
	err := h.collection.FindOne(ctx, bson.M{"is_active": true}).Decode(&token)

	if err == mongo.ErrNoDocuments {
		c.JSON(http.StatusNotFound, gin.H{"error": "No active token found"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve token: " + err.Error()})
		return
	}

	// Return token value (this should be used carefully, only by admin)
	c.JSON(http.StatusOK, gin.H{
		"access_token": token.AccessToken,
		"token_type":   token.TokenType,
		"is_active":    token.IsActive,
	})
}

// DeleteToken deactivates the Facebook access token
// DELETE /api/admin/facebook/token
func (h *FacebookHandler) DeleteToken(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := h.collection.UpdateMany(
		ctx,
		bson.M{"is_active": true},
		bson.M{"$set": bson.M{"is_active": false, "updated_at": time.Now()}},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete token: " + err.Error()})
		return
	}

	if result.ModifiedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No active token found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Token deactivated successfully",
	})
}

