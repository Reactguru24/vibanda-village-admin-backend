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

// GetEvents godoc
// @Summary Get all events
// @Description Retrieve a list of all events with pagination
// @Tags events
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search term"
// @Param status query string false "Filter by status"
// @Success 200 {object} PaginatedResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /events [get]
func GetEvents(c *gin.Context) {
	page := parseIntParam(c.Query("page"), 1)
	limit := parseIntParam(c.Query("limit"), 10)
	search := c.Query("search")
	statusFilter := c.Query("status")

	collection := database.DB.Collection("events")
	ctx := context.Background()

	// Build filter
	filter := bson.M{}
	if search != "" {
		filter["$or"] = []bson.M{
			{"title": bson.M{"$regex": search, "$options": "i"}},
			{"description": bson.M{"$regex": search, "$options": "i"}},
		}
	}
	if statusFilter != "" {
		filter["published"] = statusFilter == "published"
	}

	// Get total count
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to count events"})
		return
	}

	// Get paginated results
	opts := options.Find()
	opts.SetSkip(int64((page - 1) * limit))
	opts.SetLimit(int64(limit))
	opts.SetSort(bson.M{"created_at": -1})

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch events"})
		return
	}
	defer cursor.Close(ctx)

	var events []models.Event
	if err = cursor.All(ctx, &events); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to decode events"})
		return
	}

	// Convert to response format
	var eventResponses []models.EventResponse
	for _, event := range events {
		eventResponses = append(eventResponses, event.ToResponse())
	}

	response := PaginatedResponse{
		Data:       eventResponses,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: (total + int64(limit) - 1) / int64(limit),
	}

	c.JSON(http.StatusOK, response)
}

// GetEvent godoc
// @Summary Get event by ID
// @Description Retrieve a specific event by ID
// @Tags events
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Event ID"
// @Success 200 {object} models.EventResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /events/{id} [get]
func GetEvent(c *gin.Context) {
	id := c.Param("id")
	eventObjectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid event ID"})
		return
	}

	collection := database.DB.Collection("events")
	ctx := context.Background()

	var event models.Event
	err = collection.FindOne(ctx, bson.M{"_id": eventObjectID}).Decode(&event)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Event not found"})
		return
	}

	c.JSON(http.StatusOK, event.ToResponse())
}

// CreateEvent godoc
// @Summary Create a new event
// @Description Create a new event
// @Tags events
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.CreateEventRequest true "Event data"
// @Success 201 {object} models.EventResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /events [post]
func CreateEvent(c *gin.Context) {
	var req models.CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	collection := database.DB.Collection("events")
	ctx := context.Background()

	now := time.Now()
	event := models.Event{
		ID:          primitive.NewObjectID(),
		Title:       req.Title,
		Description: req.Description,
		Date:        req.Date,
		Location:    req.Location,
		Capacity:    req.Capacity,
		Featured:    req.Featured,
		Published:   req.Published,
		ImageURL:    req.ImageURL,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	_, err := collection.InsertOne(ctx, event)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create event"})
		return
	}

	c.JSON(http.StatusCreated, event.ToResponse())
}

// UpdateEvent godoc
// @Summary Update event
// @Description Update an existing event
// @Tags events
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Event ID"
// @Param request body models.UpdateEventRequest true "Event update data"
// @Success 200 {object} models.EventResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /events/{id} [put]
func UpdateEvent(c *gin.Context) {
	id := c.Param("id")
	eventObjectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid event ID"})
		return
	}

	var req models.UpdateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	collection := database.DB.Collection("events")
	ctx := context.Background()

	var event models.Event
	err = collection.FindOne(ctx, bson.M{"_id": eventObjectID}).Decode(&event)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Event not found"})
		return
	}

	// Update fields
	if req.Title != "" {
		event.Title = req.Title
	}
	if req.Description != "" {
		event.Description = req.Description
	}
	if req.Date != "" {
		event.Date = req.Date
	}
	if req.Location != "" {
		event.Location = req.Location
	}
	if req.Capacity > 0 {
		event.Capacity = req.Capacity
	}
	if req.ImageURL != "" {
		event.ImageURL = req.ImageURL
	}
	if req.Featured != nil {
		event.Featured = *req.Featured
	}
	if req.Published != nil {
		event.Published = *req.Published
	}

	event.UpdatedAt = time.Now()

	update := bson.M{"$set": bson.M{
		"title":       event.Title,
		"description": event.Description,
		"date":        event.Date,
		"location":    event.Location,
		"capacity":    event.Capacity,
		"image_url":   event.ImageURL,
		"featured":    event.Featured,
		"published":   event.Published,
		"updated_at":  event.UpdatedAt,
	}}

	_, err = collection.UpdateOne(ctx, bson.M{"_id": eventObjectID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update event"})
		return
	}

	c.JSON(http.StatusOK, event.ToResponse())
}

// DeleteEvent godoc
// @Summary Delete event
// @Description Delete an event
// @Tags events
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Event ID"
// @Success 204 {object} nil
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /events/{id} [delete]
func DeleteEvent(c *gin.Context) {
	id := c.Param("id")
	eventObjectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid event ID"})
		return
	}

	collection := database.DB.Collection("events")
	ctx := context.Background()

	var event models.Event
	err = collection.FindOne(ctx, bson.M{"_id": eventObjectID}).Decode(&event)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Event not found"})
		return
	}

	_, err = collection.DeleteOne(ctx, bson.M{"_id": eventObjectID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to delete event"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
