package handlers

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
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

// ProductHandler handles product-related HTTP requests
type ProductHandler struct {
	db        *database.Client
	validator *validator.Validate
}

// NewProductHandler creates a new ProductHandler
func NewProductHandler(db *database.Client) *ProductHandler {
	return &ProductHandler{
		db:        db,
		validator: validator.New(),
	}
}

// CreateProduct creates a new product (Admin only)
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	// Get user from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userRole, exists := c.Get("user_role")
	if !exists || userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	var req models.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Validate request
	if err := h.validator.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate category
	if !models.IsValidCategory(req.Category) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category"})
		return
	}

	// Convert user ID to ObjectID
	adminID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Create product
	product := models.Product{
		ID:            primitive.NewObjectID(),
		Name:          req.Name,
		Price:         req.Price,
		Category:      models.ProductCategory(req.Category),
		Description:   req.Description,
		Specification: req.Specification,
		Material:      req.Material,
		InStock:       req.InStock,
		CreatedBy:     adminID,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Insert into database
	collection := h.db.GetCollection("products")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = collection.InsertOne(ctx, product)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Product created successfully",
		"product": product.ToResponseWithBaseURL(getBaseURL(c)),
	})
}

// GetProducts retrieves all products with pagination
func (h *ProductHandler) GetProducts(c *gin.Context) {
	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	category := c.Query("category")
	inStock := c.Query("in_stock")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Build filter
	filter := bson.M{}
	if category != "" && models.IsValidCategory(category) {
		filter["category"] = category
	}
	if inStock != "" {
		if inStock == "true" {
			filter["in_stock"] = true
		} else if inStock == "false" {
			filter["in_stock"] = false
		}
	}

	collection := h.db.GetCollection("products")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Count total documents
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count products"})
		return
	}

	// Find products with pagination
	skip := (page - 1) * limit
	findOptions := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
		return
	}
	defer cursor.Close(ctx)

	var products []models.Product
	if err = cursor.All(ctx, &products); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode products"})
		return
	}

	// Convert to response format with full image URLs
	baseURL := getBaseURL(c)
	var productResponses []models.ProductResponse
	for _, product := range products {
		productResponses = append(productResponses, product.ToResponseWithBaseURL(baseURL))
	}

	response := models.ProductListResponse{
		Products: productResponses,
		Total:    total,
		Page:     page,
		Limit:    limit,
	}

	c.JSON(http.StatusOK, response)
}

// GetProduct retrieves a single product by ID
func (h *ProductHandler) GetProduct(c *gin.Context) {
	productID := c.Param("id")

	objID, err := primitive.ObjectIDFromHex(productID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	collection := h.db.GetCollection("products")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var product models.Product
	err = collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&product)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch product"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"product": product.ToResponseWithBaseURL(getBaseURL(c))})
}

// UpdateProduct updates an existing product (Admin only)
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	// Check admin access
	userRole, exists := c.Get("user_role")
	if !exists || userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	productID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(productID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	var req models.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Validate request
	if err := h.validator.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Build update document
	update := bson.M{"$set": bson.M{"updated_at": time.Now()}}

	if req.Name != nil {
		update["$set"].(bson.M)["name"] = *req.Name
	}
	if req.Price != nil {
		update["$set"].(bson.M)["price"] = *req.Price
	}
	if req.Category != nil {
		if !models.IsValidCategory(*req.Category) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category"})
			return
		}
		update["$set"].(bson.M)["category"] = *req.Category
	}
	if req.Description != nil {
		update["$set"].(bson.M)["description"] = *req.Description
	}
	if req.Specification != nil {
		update["$set"].(bson.M)["specification"] = *req.Specification
	}
	if req.Material != nil {
		update["$set"].(bson.M)["material"] = *req.Material
	}
	if req.InStock != nil {
		update["$set"].(bson.M)["in_stock"] = *req.InStock
	}

	collection := h.db.GetCollection("products")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product"})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	// Fetch updated product
	var updatedProduct models.Product
	err = collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&updatedProduct)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch updated product"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Product updated successfully",
		"product": updatedProduct.ToResponseWithBaseURL(getBaseURL(c)),
	})
}

// DeleteProduct deletes a product (Admin only)
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	// Check admin access
	userRole, exists := c.Get("user_role")
	if !exists || userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	productID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(productID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	collection := h.db.GetCollection("products")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if product exists and get image URL for cleanup
	var product models.Product
	err = collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&product)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch product"})
		return
	}

	// Delete product from database
	result, err := collection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product"})
		return
	}

	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	// Clean up image file if exists
	if product.ImageURL != "" {
		// Extract filename from URL and delete file
		filename := filepath.Base(product.ImageURL)
		imagePath := filepath.Join("uploads", "products", filename)
		os.Remove(imagePath) // Ignore error if file doesn't exist
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}

// UploadProductImage handles product image upload (Admin only)
func (h *ProductHandler) UploadProductImage(c *gin.Context) {
	// Check admin access
	userRole, exists := c.Get("user_role")
	if !exists || userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	productID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(productID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
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
	uploadDir := "uploads/products"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
		return
	}

	// Generate unique filename
	filename := fmt.Sprintf("%s_%d%s", productID, time.Now().Unix(), filepath.Ext(header.Filename))
	filePath := filepath.Join(uploadDir, filename)

	// Save file
	if err := saveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image"})
		return
	}

	// Update product with image URL
	imageURL := fmt.Sprintf("/uploads/products/%s", filename)
	collection := h.db.GetCollection("products")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"image_url":  imageURL,
			"updated_at": time.Now(),
		},
	}

	result, err := collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		// Clean up uploaded file if database update fails
		os.Remove(filePath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product with image", "details": err.Error()})
		return
	}

	if result.MatchedCount == 0 {
		// Clean up uploaded file if product not found
		os.Remove(filePath)
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	if result.ModifiedCount == 0 {
		// Product found but update didn't modify anything - this should be rare but handle it
		// Don't remove file in this case as it might be a duplicate update
		c.JSON(http.StatusOK, gin.H{
			"message":   "Image uploaded but product already has this image URL",
			"image_url": imageURL,
			"warning":   "No database update was needed",
		})
		return
	}

	// Verify the update was successful by fetching the updated product
	var updatedProduct models.Product
	err = collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&updatedProduct)
	if err != nil {
		// Log error but still return success since UpdateOne succeeded
		// The image is saved and database update was confirmed
		fullImageURL := imageURL
		baseURL := getBaseURL(c)
		if baseURL != "" && imageURL != "" && !strings.HasPrefix(imageURL, "http://") && !strings.HasPrefix(imageURL, "https://") {
			if strings.HasPrefix(imageURL, "/") {
				fullImageURL = baseURL + imageURL
			} else {
				fullImageURL = baseURL + "/" + imageURL
			}
		}
		c.JSON(http.StatusOK, gin.H{
			"message":   "Image uploaded successfully",
			"image_url": fullImageURL,
		})
		return
	}

	// Verify image_url was actually set in the database
	if updatedProduct.ImageURL == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Image uploaded but database update verification failed",
		})
		return
	}

	// Return full URL for the image
	fullImageURL := updatedProduct.ImageURL
	baseURL := getBaseURL(c)
	if baseURL != "" && fullImageURL != "" && !strings.HasPrefix(fullImageURL, "http://") && !strings.HasPrefix(fullImageURL, "https://") {
		if strings.HasPrefix(fullImageURL, "/") {
			fullImageURL = baseURL + fullImageURL
		} else {
			fullImageURL = baseURL + "/" + fullImageURL
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Image uploaded successfully",
		"image_url": fullImageURL,
	})
}

// GetCategories returns all available product categories
func (h *ProductHandler) GetCategories(c *gin.Context) {
	categories := models.GetValidCategories()
	var categoryStrings []string
	for _, category := range categories {
		categoryStrings = append(categoryStrings, string(category))
	}

	response := models.CategoriesResponse{
		Categories: categoryStrings,
	}

	c.JSON(http.StatusOK, response)
}

// Helper functions

// getBaseURL extracts the base URL from the request context
func getBaseURL(c *gin.Context) string {
	scheme := "http"
	if c.GetHeader("X-Forwarded-Proto") == "https" || c.Request.TLS != nil {
		scheme = "https"
	}
	host := c.GetHeader("X-Forwarded-Host")
	if host == "" {
		host = c.Request.Host
	}
	if host == "" {
		host = "localhost:8080" // fallback
	}
	return fmt.Sprintf("%s://%s", scheme, host)
}

// isValidImageType checks if the file has a valid image extension
func isValidImageType(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	validExts := []string{".jpg", ".jpeg", ".png", ".gif"}

	for _, validExt := range validExts {
		if ext == validExt {
			return true
		}
	}
	return false
}

// saveUploadedFile saves the uploaded file to the specified path
func saveUploadedFile(file multipart.File, dst string) error {
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	return err
}
