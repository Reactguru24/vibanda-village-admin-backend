package handlers

import (
	"context"
	"net/http"
	"time"
	"vibanda-village-backend/internal/database"
	"vibanda-village-backend/internal/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetProducts godoc
// @Summary Get all products
// @Description Retrieve a list of all products with pagination
// @Tags products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search term"
// @Param category query string false "Filter by category"
// @Param status query string false "Filter by status"
// @Success 200 {object} PaginatedResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /products [get]
func GetProducts(c *gin.Context) {
	page := parseIntParam(c.Query("page"), 1)
	limit := parseIntParam(c.Query("limit"), 10)
	search := c.Query("search")
	categoryFilter := c.Query("category")
	statusFilter := c.Query("status")

	collection := database.DB.Collection("products")
	ctx := context.Background()

	// Build filter
	filter := bson.M{}
	if search != "" {
		filter["$or"] = []bson.M{
			{"name": bson.M{"$regex": search, "$options": "i"}},
			{"description": bson.M{"$regex": search, "$options": "i"}},
		}
	}
	if categoryFilter != "" {
		filter["category"] = categoryFilter
	}
	if statusFilter != "" {
		filter["available"] = statusFilter == "active"
	}

	// Get total count
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to count products"})
		return
	}

	// Get paginated results
	opts := options.Find()
	opts.SetSkip(int64((page - 1) * limit))
	opts.SetLimit(int64(limit))
	opts.SetSort(bson.M{"created_at": -1})

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch products"})
		return
	}
	defer cursor.Close(ctx)

	var products []models.Product
	if err = cursor.All(ctx, &products); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to decode products"})
		return
	}

	// Convert to response format
	var productResponses []models.ProductResponse
	for _, product := range products {
		productResponses = append(productResponses, product.ToResponse())
	}

	response := PaginatedResponse{
		Data:       productResponses,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: (total + int64(limit) - 1) / int64(limit),
	}

	c.JSON(http.StatusOK, response)
}

// GetProduct godoc
// @Summary Get product by ID
// @Description Retrieve a specific product by ID
// @Tags products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Product ID"
// @Success 200 {object} models.ProductResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /products/{id} [get]
func GetProduct(c *gin.Context) {
	id := c.Param("id")
	productObjectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid product ID"})
		return
	}

	collection := database.DB.Collection("products")
	ctx := context.Background()

	var product models.Product
	err = collection.FindOne(ctx, bson.M{"_id": productObjectID}).Decode(&product)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Product not found"})
		return
	}

	c.JSON(http.StatusOK, product.ToResponse())
}

// CreateProduct godoc
// @Summary Create a new product
// @Description Create a new product
// @Tags products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.CreateProductRequest true "Product data"
// @Success 201 {object} models.ProductResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /products [post]
func CreateProduct(c *gin.Context) {
	var req models.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	collection := database.DB.Collection("products")
	ctx := context.Background()

	now := time.Now()
	product := models.Product{
		ID:          primitive.NewObjectID(),
		Name:        req.Name,
		Category:    req.Category,
		Subcategory: req.Subcategory,
		Price:       req.Price,
		Stock:       req.Stock,
		Description: req.Description,
		ImageURL:    req.ImageURL,
		Popular:     req.Popular,
		New:         req.New,
		Available:   req.Available,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	_, err := collection.InsertOne(ctx, product)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create product"})
		return
	}

	c.JSON(http.StatusCreated, product.ToResponse())
}

// UpdateProduct godoc
// @Summary Update product
// @Description Update an existing product
// @Tags products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Product ID"
// @Param request body models.UpdateProductRequest true "Product update data"
// @Success 200 {object} models.ProductResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /products/{id} [put]
func UpdateProduct(c *gin.Context) {
	id := c.Param("id")
	productObjectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid product ID"})
		return
	}

	var req models.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	collection := database.DB.Collection("products")
	ctx := context.Background()

	var product models.Product
	err = collection.FindOne(ctx, bson.M{"_id": productObjectID}).Decode(&product)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Product not found"})
		return
	}

	// Update fields
	if req.Name != "" {
		product.Name = req.Name
	}
	if req.Description != "" {
		product.Description = req.Description
	}
	if req.Price > 0 {
		product.Price = req.Price
	}
	if req.Category != "" {
		product.Category = req.Category
	}
	if req.Subcategory != "" {
		product.Subcategory = req.Subcategory
	}
	if req.ImageURL != "" {
		product.ImageURL = req.ImageURL
	}
	if req.Stock >= 0 {
		product.Stock = req.Stock
	}
	if req.Popular != nil {
		product.Popular = *req.Popular
	}
	if req.New != nil {
		product.New = *req.New
	}
	if req.Available != nil {
		product.Available = *req.Available
	}

	product.UpdatedAt = time.Now()

	update := bson.M{"$set": bson.M{
		"name":        product.Name,
		"description": product.Description,
		"price":       product.Price,
		"category":    product.Category,
		"subcategory": product.Subcategory,
		"image_url":   product.ImageURL,
		"stock":       product.Stock,
		"popular":     product.Popular,
		"new":         product.New,
		"available":   product.Available,
		"updated_at":  product.UpdatedAt,
	}}

	_, err = collection.UpdateOne(ctx, bson.M{"_id": productObjectID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update product"})
		return
	}

	c.JSON(http.StatusOK, product.ToResponse())
}

// DeleteProduct godoc
// @Summary Delete product
// @Description Delete a product
// @Tags products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Product ID"
// @Success 204 {object} nil
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /products/{id} [delete]
func DeleteProduct(c *gin.Context) {
	id := c.Param("id")
	productObjectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid product ID"})
		return
	}

	collection := database.DB.Collection("products")
	ctx := context.Background()

	var product models.Product
	err = collection.FindOne(ctx, bson.M{"_id": productObjectID}).Decode(&product)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Product not found"})
		return
	}

	_, err = collection.DeleteOne(ctx, bson.M{"_id": productObjectID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to delete product"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
