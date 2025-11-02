package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"ecommerce-backend/internal/database"
	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// OrderHandler handles order-related HTTP requests
type OrderHandler struct {
	db         *database.Client
	validator  *validator.Validate
	jwtManager interface{} // Will be set during initialization
}

// NewOrderHandler creates a new OrderHandler
func NewOrderHandler(db *database.Client, jwtManager interface{}) *OrderHandler {
	return &OrderHandler{
		db:         db,
		validator:  validator.New(),
		jwtManager: jwtManager,
	}
}

// CreateOrder handles order creation (Public - authentication optional)
// If user provides a valid token, user_id will be saved; otherwise, guest order is created
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	// Try to get user ID if authenticated
	// Since this is a public route, manually check for token in Authorization header
	var objUserID *primitive.ObjectID
	
	// Check for Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenString := authHeader[7:]
		// Try to validate token using JWT manager
		if jwtMgr, ok := h.jwtManager.(*utils.JWTManager); ok {
			claims, err := jwtMgr.ValidateToken(tokenString)
			if err == nil && claims != nil {
				// Token is valid, extract user_id
				tempUserID, err := primitive.ObjectIDFromHex(claims.UserID)
				if err == nil {
					objUserID = &tempUserID
				}
			}
		}
	}
	
	// If not authenticated, objUserID will be nil (guest order)

	var req models.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload", "details": err.Error()})
		return
	}

	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
		return
	}

	// Fetch products and build order items
	collection := h.db.GetCollection("products")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var orderItems []models.OrderItem

	for _, itemReq := range req.Items {
		var product models.Product
		err := collection.FindOne(ctx, bson.M{"_id": itemReq.ProductID}).Decode(&product)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Product not found", "product_id": itemReq.ProductID.Hex()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch product"})
			return
		}

		// Check stock
		if !product.InStock {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Product out of stock", "product_id": product.ID.Hex(), "product_name": product.Name})
			return
		}

		itemSubtotal := product.Price * float64(itemReq.Quantity)
		orderItems = append(orderItems, models.OrderItem{
			ProductID:   product.ID,
			ProductName: product.Name,
			Quantity:    itemReq.Quantity,
			Price:       product.Price,
			Subtotal:    itemSubtotal,
		})
	}

	// Use values sent by frontend - no calculation, just save what they send
	subtotal := req.Subtotal
	shippingCost := req.ShippingCost
	tax := req.Tax
	total := req.Total

	// Create order
	orderCollection := h.db.GetCollection("orders")
	ctx2, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel2()
	
	// Get guest name from shipping address if not authenticated
	guestName := ""
	if objUserID == nil {
		guestName = req.ShippingAddress.FullName
	}
	
	// Generate custom order ID (broshopbd_000001)
	// Get the last order to determine the next sequence number
	var lastOrder models.Order
	findOptions := options.FindOne().SetSort(bson.D{{Key: "_id", Value: -1}})
	lastErr := orderCollection.FindOne(ctx2, bson.M{}, findOptions).Decode(&lastOrder)
	
	orderSequence := 1
	if lastErr == nil && lastOrder.OrderID != "" {
		// Extract number from last order ID (e.g., "broshopbd_000001" -> 1)
		// Simple parsing: take the part after underscore
		parts := len(lastOrder.OrderID)
		if parts > 11 { // "broshopbd_" is 10 chars
			lastNumStr := lastOrder.OrderID[11:] // Get everything after "broshopbd_"
			if lastNum, parseErr := strconv.Atoi(lastNumStr); parseErr == nil {
				orderSequence = lastNum + 1
			}
		}
	}
	
	// Format: broshopbd_000001
	customOrderID := fmt.Sprintf("broshopbd_%06d", orderSequence)
	
	order := models.Order{
		ID:              primitive.NewObjectID(),
		OrderID:         customOrderID,
		UserID:          objUserID, // Will be nil for guest orders
		GuestName:       guestName,  // Set for guest orders
		Items:           orderItems,
		ShippingAddress: req.ShippingAddress,
		Subtotal:        subtotal,
		ShippingCost:    shippingCost,
		Tax:             tax,
		Total:           total,
		Status:          models.OrderStatusPending,
		PaymentMethod:   req.PaymentMethod,
		PaymentStatus:   "pending",
		OrderNotes:      req.OrderNotes,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	_, err := orderCollection.InsertOne(ctx, order)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Order created successfully",
		"order":   order.ToResponse(),
	})
}

// GetOrders handles listing orders (Admin: all orders, User: own orders) - Requires authentication
func (h *OrderHandler) GetOrders(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	userRole, _ := c.Get("user_role")

	collection := h.db.GetCollection("orders")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	skip := (page - 1) * limit

	// Build filter
	filter := bson.M{}
	if userRole != "admin" {
		// Users can only see their own orders
		objUserID, err := primitive.ObjectIDFromHex(userID.(string))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}
		filter["user_id"] = objUserID
	}

	// Status filter
	if status := c.Query("status"); status != "" {
		filter["status"] = status
	}

	// Date filters
	dateFilter := bson.M{}
	
	// Start date filter (from_date)
	if fromDate := c.Query("from_date"); fromDate != "" {
		// Expected format: 2025-11-01 or 2025-11-01T00:00:00Z
		parsedDate, err := time.Parse("2006-01-02", fromDate)
		if err != nil {
			// Try RFC3339 format
			parsedDate, err = time.Parse(time.RFC3339, fromDate)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid from_date format. Use YYYY-MM-DD or RFC3339"})
				return
			}
		}
		dateFilter["$gte"] = parsedDate
	}
	
	// End date filter (to_date)
	if toDate := c.Query("to_date"); toDate != "" {
		parsedDate, err := time.Parse("2006-01-02", toDate)
		if err != nil {
			parsedDate, err = time.Parse(time.RFC3339, toDate)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid to_date format. Use YYYY-MM-DD or RFC3339"})
				return
			}
		}
		// Add one day to include the entire end date
		endOfDay := parsedDate.Add(24 * time.Hour)
		dateFilter["$lt"] = endOfDay
	}
	
	// Apply date filter if any date parameters were provided
	if len(dateFilter) > 0 {
		filter["created_at"] = dateFilter
	}

	// Find options
	findOptions := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetSkip(int64(skip)).
		SetLimit(int64(limit))

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders"})
		return
	}
	defer cursor.Close(ctx)

	var orders []models.Order
	if err = cursor.All(ctx, &orders); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode orders"})
		return
	}

	// Get total count
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count orders"})
		return
	}

	// Convert to response
	var orderResponses []models.OrderResponse
	statusCount := make(map[string]int)
	for _, order := range orders {
		orderResp := order.ToResponse()
		orderResponses = append(orderResponses, orderResp)
		// Count orders by status
		statusCount[orderResp.Status] = statusCount[orderResp.Status] + 1
	}

	response := models.OrderListResponse{
		Orders:      orderResponses,
		Total:       int(total),
		Page:        page,
		Limit:       limit,
		StatusCount: statusCount,
	}

	c.JSON(http.StatusOK, response)
}

// GetOrder handles fetching a single order (User: own orders only, Admin: any order)
// Accepts custom order_id (e.g., broshopbd_000001) or MongoDB ObjectID
func (h *OrderHandler) GetOrder(c *gin.Context) {
	orderID := c.Param("id")
	
	userID, _ := c.Get("user_id")
	userRole, _ := c.Get("user_role")

	collection := h.db.GetCollection("orders")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Try to find by custom order_id first (e.g., broshopbd_000001)
	var order models.Order
	err := collection.FindOne(ctx, bson.M{"order_id": orderID}).Decode(&order)
	
	// If not found, try MongoDB ObjectID (for backward compatibility)
	if err == mongo.ErrNoDocuments {
		objID, parseErr := primitive.ObjectIDFromHex(orderID)
		if parseErr == nil {
			err = collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&order)
		}
	}
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch order"})
		return
	}

	// Check permission: users can only see their own orders
	if userRole != "admin" {
		objUserID, err := primitive.ObjectIDFromHex(userID.(string))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}
		// Check if order belongs to user
		if order.UserID == nil || *order.UserID != objUserID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"order": order.ToResponse()})
}

// UpdateOrder handles order updates (Admin: full update, User: limited update)
func (h *OrderHandler) UpdateOrder(c *gin.Context) {
	orderID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	userID, _ := c.Get("user_id")
	userRole, _ := c.Get("user_role")

	var req models.UpdateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload", "details": err.Error()})
		return
	}

	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
		return
	}

	collection := h.db.GetCollection("orders")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Fetch existing order
	var order models.Order
	err = collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&order)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch order"})
		return
	}

	// Check permission
	if userRole != "admin" {
		objUserID, err := primitive.ObjectIDFromHex(userID.(string))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}
		// Check if order belongs to user
		if order.UserID == nil || *order.UserID != objUserID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}

		// Users can only update shipping address and only if order is pending
		if order.Status != models.OrderStatusPending {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Can only update address for pending orders"})
			return
		}

		// Users can only update shipping address
		if req.Status != nil || req.AdminNotes != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Users cannot update status or admin notes"})
			return
		}
	}

	// Build update
	update := bson.M{
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	if req.Status != nil && userRole == "admin" {
		update["$set"].(bson.M)["status"] = *req.Status

		// Set timestamp for shipped/delivered
		now := time.Now()
		if *req.Status == models.OrderStatusShipped && order.ShippedAt == nil {
			update["$set"].(bson.M)["shipped_at"] = now
		}
		if *req.Status == models.OrderStatusDelivered && order.DeliveredAt == nil {
			update["$set"].(bson.M)["delivered_at"] = now
		}
	}

	if req.AdminNotes != nil && userRole == "admin" {
		update["$set"].(bson.M)["admin_notes"] = *req.AdminNotes
	}

	if req.ShippingAddress != nil {
		update["$set"].(bson.M)["shipping_address"] = *req.ShippingAddress
	}

	result, err := collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order"})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	// Fetch updated order
	var updatedOrder models.Order
	collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&updatedOrder)

	c.JSON(http.StatusOK, gin.H{
		"message": "Order updated successfully",
		"order":   updatedOrder.ToResponse(),
	})
}

// UpdateOrderStatus handles order status updates (Admin only)
func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	userRole, exists := c.Get("user_role")
	if !exists || userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	orderID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	var req models.UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload", "details": err.Error()})
		return
	}

	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
		return
	}

	collection := h.db.GetCollection("orders")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Fetch existing order to check current status
	var order models.Order
	err = collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&order)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch order"})
		return
	}

	// Build update
	update := bson.M{
		"$set": bson.M{
			"status":     req.Status,
			"updated_at": time.Now(),
		},
	}

	// Set timestamp for shipped/delivered
	now := time.Now()
	if req.Status == models.OrderStatusShipped && order.ShippedAt == nil {
		update["$set"].(bson.M)["shipped_at"] = now
	}
	if req.Status == models.OrderStatusDelivered && order.DeliveredAt == nil {
		update["$set"].(bson.M)["delivered_at"] = now
	}

	if req.AdminNotes != nil {
		update["$set"].(bson.M)["admin_notes"] = *req.AdminNotes
	}

	result, err := collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order status"})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	// Fetch updated order
	var updatedOrder models.Order
	collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&updatedOrder)

	c.JSON(http.StatusOK, gin.H{
		"message": "Order status updated successfully",
		"order":   updatedOrder.ToResponse(),
	})
}

