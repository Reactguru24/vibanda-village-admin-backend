package handlers

import (
	"context"
	"net/http"
	"time"
	"vibanda-village-admin-backend/internal/database"
	"vibanda-village-admin-backend/internal/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetReservations godoc
// @Summary Get all reservations
// @Description Retrieve a list of all reservations with pagination
// @Tags reservations
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
// @Router /reservations [get]
func GetReservations(c *gin.Context) {
	page := parseIntParam(c.Query("page"), 1)
	limit := parseIntParam(c.Query("limit"), 10)
	search := c.Query("search")
	statusFilter := c.Query("status")

	collection := database.DB.Collection("reservations")
	ctx := context.Background()

	// Build filter
	filter := bson.M{}
	if search != "" {
		filter["$or"] = []bson.M{
			{"customer_name": bson.M{"$regex": search, "$options": "i"}},
			{"customer_email": bson.M{"$regex": search, "$options": "i"}},
		}
	}
	if statusFilter != "" {
		filter["status"] = statusFilter
	}

	// Get total count
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to count reservations"})
		return
	}

	// Get paginated results
	opts := options.Find()
	opts.SetSkip(int64((page - 1) * limit))
	opts.SetLimit(int64(limit))
	opts.SetSort(bson.M{"created_at": -1})

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch reservations"})
		return
	}
	defer cursor.Close(ctx)

	var reservations []models.Reservation
	if err = cursor.All(ctx, &reservations); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to decode reservations"})
		return
	}

	// Convert to response format
	var reservationResponses []models.ReservationResponse
	for _, reservation := range reservations {
		reservationResponses = append(reservationResponses, reservation.ToResponse())
	}

	response := PaginatedResponse{
		Data:       reservationResponses,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: (total + int64(limit) - 1) / int64(limit),
	}

	c.JSON(http.StatusOK, response)
}

// GetReservation godoc
// @Summary Get reservation by ID
// @Description Retrieve a specific reservation by ID
// @Tags reservations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Reservation ID"
// @Success 200 {object} models.ReservationResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /reservations/{id} [get]
func GetReservation(c *gin.Context) {
	id := c.Param("id")
	reservationObjectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid reservation ID"})
		return
	}

	collection := database.DB.Collection("reservations")
	ctx := context.Background()

	var reservation models.Reservation
	err = collection.FindOne(ctx, bson.M{"_id": reservationObjectID}).Decode(&reservation)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Reservation not found"})
		return
	}

	c.JSON(http.StatusOK, reservation.ToResponse())
}

// CreateReservation godoc
// @Summary Create a new reservation
// @Description Create a new reservation
// @Tags reservations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.CreateReservationRequest true "Reservation data"
// @Success 201 {object} models.ReservationResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /reservations [post]
func CreateReservation(c *gin.Context) {
	var req models.CreateReservationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	collection := database.DB.Collection("reservations")
	ctx := context.Background()

	now := time.Now()
	reservation := models.Reservation{
		ID:              primitive.NewObjectID(),
		CustomerName:    req.CustomerName,
		CustomerEmail:   req.CustomerEmail,
		CustomerPhone:   req.CustomerPhone,
		Date:            req.Date,
		Time:            req.Time,
		Guests:          req.Guests,
		Status:          models.ReservationStatusPending,
		SpecialRequests: req.SpecialRequests,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	_, err := collection.InsertOne(ctx, reservation)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create reservation"})
		return
	}

	c.JSON(http.StatusCreated, reservation.ToResponse())
}

// UpdateReservation godoc
// @Summary Update reservation
// @Description Update an existing reservation
// @Tags reservations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Reservation ID"
// @Param request body models.UpdateReservationRequest true "Reservation update data"
// @Success 200 {object} models.ReservationResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /reservations/{id} [put]
func UpdateReservation(c *gin.Context) {
	id := c.Param("id")
	reservationObjectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid reservation ID"})
		return
	}

	var req models.UpdateReservationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	collection := database.DB.Collection("reservations")
	ctx := context.Background()

	var reservation models.Reservation
	err = collection.FindOne(ctx, bson.M{"_id": reservationObjectID}).Decode(&reservation)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Reservation not found"})
		return
	}

	// Update fields
	if req.CustomerName != "" {
		reservation.CustomerName = req.CustomerName
	}
	if req.CustomerEmail != "" {
		reservation.CustomerEmail = req.CustomerEmail
	}
	if req.CustomerPhone != "" {
		reservation.CustomerPhone = req.CustomerPhone
	}
	if req.Date != "" {
		reservation.Date = req.Date
	}
	if req.Time != "" {
		reservation.Time = req.Time
	}
	if req.Guests > 0 {
		reservation.Guests = req.Guests
	}
	if req.Status != "" {
		reservation.Status = req.Status
	}
	if req.SpecialRequests != "" {
		reservation.SpecialRequests = req.SpecialRequests
	}

	reservation.UpdatedAt = time.Now()

	update := bson.M{"$set": bson.M{
		"customer_name":    reservation.CustomerName,
		"customer_email":   reservation.CustomerEmail,
		"customer_phone":   reservation.CustomerPhone,
		"date":             reservation.Date,
		"time":             reservation.Time,
		"guests":           reservation.Guests,
		"status":           reservation.Status,
		"special_requests": reservation.SpecialRequests,
		"updated_at":       reservation.UpdatedAt,
	}}

	_, err = collection.UpdateOne(ctx, bson.M{"_id": reservationObjectID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update reservation"})
		return
	}

	c.JSON(http.StatusOK, reservation.ToResponse())
}

// DeleteReservation godoc
// @Summary Delete reservation
// @Description Delete a reservation
// @Tags reservations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Reservation ID"
// @Success 204 {object} nil
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /reservations/{id} [delete]
func DeleteReservation(c *gin.Context) {
	id := c.Param("id")
	reservationObjectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid reservation ID"})
		return
	}

	collection := database.DB.Collection("reservations")
	ctx := context.Background()

	var reservation models.Reservation
	err = collection.FindOne(ctx, bson.M{"_id": reservationObjectID}).Decode(&reservation)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Reservation not found"})
		return
	}

	_, err = collection.DeleteOne(ctx, bson.M{"_id": reservationObjectID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to delete reservation"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
