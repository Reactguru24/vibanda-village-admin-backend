package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"
	"vibanda-village-admin-backend/internal/database"
	"vibanda-village-admin-backend/internal/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// generateOrderNumber generates a unique order number
func generateOrderNumber() string {
	return fmt.Sprintf("ORD-%d", time.Now().Unix())
}

// GetOrders godoc
// @Summary Get all orders
// @Description Retrieve a list of all orders with pagination
// @Tags orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search term"
// @Param status query string false "Filter by status"
// @Param payment_status query string false "Filter by payment status"
// @Success 200 {object} PaginatedResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /orders [get]
func GetOrders(c *gin.Context) {
	page := parseIntParam(c.Query("page"), 1)
	limit := parseIntParam(c.Query("limit"), 10)
	search := c.Query("search")
	statusFilter := c.Query("status")
	paymentStatusFilter := c.Query("payment_status")

	collection := database.DB.Collection("orders")
	ctx := context.Background()

	// Build filter
	filter := bson.M{}
	if search != "" {
		filter["$or"] = []bson.M{
			{"order_number": bson.M{"$regex": search, "$options": "i"}},
			{"customer_name": bson.M{"$regex": search, "$options": "i"}},
		}
	}
	if statusFilter != "" {
		filter["status"] = statusFilter
	}
	if paymentStatusFilter != "" {
		filter["payment_status"] = paymentStatusFilter
	}

	// Get total count
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to count orders"})
		return
	}

	// Get paginated results
	opts := options.Find()
	opts.SetSkip(int64((page - 1) * limit))
	opts.SetLimit(int64(limit))
	opts.SetSort(bson.M{"created_at": -1})

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch orders"})
		return
	}
	defer cursor.Close(ctx)

	var orders []models.Order
	if err = cursor.All(ctx, &orders); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to decode orders"})
		return
	}

	// Convert to response format
	var orderResponses []models.OrderResponse
	for _, order := range orders {
		orderResponses = append(orderResponses, order.ToResponse())
	}

	response := PaginatedResponse{
		Data:       orderResponses,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: (total + int64(limit) - 1) / int64(limit),
	}

	c.JSON(http.StatusOK, response)
}

// GetOrder godoc
// @Summary Get order by ID
// @Description Retrieve a specific order by ID
// @Tags orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Success 200 {object} models.OrderResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /orders/{id} [get]
func GetOrder(c *gin.Context) {
	id := c.Param("id")
	orderObjectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid order ID"})
		return
	}

	collection := database.DB.Collection("orders")
	ctx := context.Background()

	var order models.Order
	err = collection.FindOne(ctx, bson.M{"_id": orderObjectID}).Decode(&order)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Order not found"})
		return
	}

	c.JSON(http.StatusOK, order.ToResponse())
}

// CreateOrder godoc
// @Summary Create a new order
// @Description Create a new order
// @Tags orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.CreateOrderRequest true "Order data"
// @Success 201 {object} models.OrderResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /orders [post]
func CreateOrder(c *gin.Context) {
	var req models.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	collection := database.DB.Collection("orders")
	ctx := context.Background()

	// Convert request items to order items
	var items []models.OrderItem
	totalAmount := 0.0
	for _, itemReq := range req.Items {
		item := models.OrderItem{
			ID:       primitive.NewObjectID(),
			Name:     itemReq.Name,
			Quantity: itemReq.Quantity,
			Price:    itemReq.Price,
		}
		items = append(items, item)
		totalAmount += itemReq.Price * float64(itemReq.Quantity)
	}

	now := time.Now()
	order := models.Order{
		ID:             primitive.NewObjectID(),
		OrderNumber:    generateOrderNumber(),
		CustomerName:   req.CustomerName,
		CustomerPhone:  req.CustomerPhone,
		CustomerEmail:  req.CustomerEmail,
		TotalAmount:    totalAmount,
		Status:         models.OrderStatusPending,
		PaymentStatus:  models.PaymentStatusPending,
		SpecialRequest: req.SpecialRequest,
		Items:          items,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	_, err := collection.InsertOne(ctx, order)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create order"})
		return
	}

	c.JSON(http.StatusCreated, order.ToResponse())
}

// UpdateOrder godoc
// @Summary Update order
// @Description Update an existing order
// @Tags orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Param request body models.UpdateOrderRequest true "Order update data"
// @Success 200 {object} models.OrderResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /orders/{id} [put]
func UpdateOrder(c *gin.Context) {
	id := c.Param("id")
	orderObjectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid order ID"})
		return
	}

	var req models.UpdateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	collection := database.DB.Collection("orders")
	ctx := context.Background()

	var order models.Order
	err = collection.FindOne(ctx, bson.M{"_id": orderObjectID}).Decode(&order)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Order not found"})
		return
	}

	// Update fields
	if req.CustomerName != "" {
		order.CustomerName = req.CustomerName
	}
	if req.CustomerPhone != "" {
		order.CustomerPhone = req.CustomerPhone
	}
	if req.CustomerEmail != "" {
		order.CustomerEmail = req.CustomerEmail
	}
	if req.Status != "" {
		order.Status = req.Status
	}
	if req.PaymentStatus != "" {
		order.PaymentStatus = req.PaymentStatus
	}
	if req.SpecialRequest != "" {
		order.SpecialRequest = req.SpecialRequest
	}

	order.UpdatedAt = time.Now()

	update := bson.M{"$set": bson.M{
		"customer_name":   order.CustomerName,
		"customer_phone":  order.CustomerPhone,
		"customer_email":  order.CustomerEmail,
		"status":          order.Status,
		"payment_status":  order.PaymentStatus,
		"special_request": order.SpecialRequest,
		"updated_at":      order.UpdatedAt,
	}}

	_, err = collection.UpdateOne(ctx, bson.M{"_id": orderObjectID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update order"})
		return
	}

	c.JSON(http.StatusOK, order.ToResponse())
}

// DeleteOrder godoc
// @Summary Delete order
// @Description Delete an order
// @Tags orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Success 204 {object} nil
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /orders/{id} [delete]
func DeleteOrder(c *gin.Context) {
	id := c.Param("id")
	orderObjectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid order ID"})
		return
	}

	collection := database.DB.Collection("orders")
	ctx := context.Background()

	var order models.Order
	err = collection.FindOne(ctx, bson.M{"_id": orderObjectID}).Decode(&order)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Order not found"})
		return
	}

	_, err = collection.DeleteOne(ctx, bson.M{"_id": orderObjectID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to delete order"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
