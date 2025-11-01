package handlers

import (
	"context"
	"net/http"
	"time"
	"vibanda-village-admin-backend/internal/database"
	"vibanda-village-admin-backend/internal/models"
	"vibanda-village-admin-backend/pkg/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetUsers godoc
// @Summary Get all users
// @Description Retrieve a list of all users with pagination
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search term"
// @Param role query string false "Filter by role"
// @Param status query string false "Filter by status"
// @Success 200 {object} PaginatedResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users [get]
func GetUsers(c *gin.Context) {
	page := parseIntParam(c.Query("page"), 1)
	limit := parseIntParam(c.Query("limit"), 10)
	search := c.Query("search")
	roleFilter := c.Query("role")
	statusFilter := c.Query("status")

	collection := database.DB.Collection("users")
	ctx := context.Background()

	// Build filter
	filter := bson.M{}
	if search != "" {
		filter["$or"] = []bson.M{
			{"name": bson.M{"$regex": search, "$options": "i"}},
			{"email": bson.M{"$regex": search, "$options": "i"}},
			{"username": bson.M{"$regex": search, "$options": "i"}},
		}
	}
	if roleFilter != "" {
		filter["role"] = roleFilter
	}
	if statusFilter != "" {
		filter["status"] = statusFilter
	}

	// Get total count
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to count users"})
		return
	}

	// Get paginated results
	opts := options.Find()
	opts.SetSkip(int64((page - 1) * limit))
	opts.SetLimit(int64(limit))
	opts.SetSort(bson.M{"created_at": -1})

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch users"})
		return
	}
	defer cursor.Close(ctx)

	var users []models.User
	if err = cursor.All(ctx, &users); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to decode users"})
		return
	}

	// Convert to response format
	var userResponses []models.UserResponse
	for _, user := range users {
		userResponses = append(userResponses, user.ToResponse())
	}

	response := PaginatedResponse{
		Data:       userResponses,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: (total + int64(limit) - 1) / int64(limit),
	}

	c.JSON(http.StatusOK, response)
}

// GetUser godoc
// @Summary Get user by ID
// @Description Retrieve a specific user by ID
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} models.UserResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{id} [get]
func GetUser(c *gin.Context) {
	id := c.Param("id")
	userObjectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid user ID"})
		return
	}

	collection := database.DB.Collection("users")
	ctx := context.Background()

	var user models.User
	err = collection.FindOne(ctx, bson.M{"_id": userObjectID}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "User not found"})
		return
	}

	c.JSON(http.StatusOK, user.ToResponse())
}

// CreateUser godoc
// @Summary Create a new user
// @Description Create a new user account (Admin can create managers and staff, Manager can create staff only)
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.RegisterRequest true "User data"
// @Success 201 {object} models.UserResponse
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users [post]
func CreateUser(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Get current user from context (set by auth middleware)
	currentUserID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	collection := database.DB.Collection("users")
	ctx := context.Background()

	currentUserObjectID, err := primitive.ObjectIDFromHex(currentUserID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid current user ID"})
		return
	}

	var currentUser models.User
	err = collection.FindOne(ctx, bson.M{"_id": currentUserObjectID}).Decode(&currentUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get current user"})
		return
	}

	// Permission checks
	if currentUser.Role == models.RoleAdmin {
		// Admin can create managers and staff, but not other admins
		if req.Role == models.RoleAdmin {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: "Admin cannot create other admins"})
			return
		}
	} else if currentUser.Role == models.RoleManager {
		// Manager can only create staff
		if req.Role != models.RoleStaff {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: "Manager can only create staff accounts"})
			return
		}
	} else {
		// Staff cannot create users
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "Staff cannot create user accounts"})
		return
	}

	// Check if user already exists
	var existingUser models.User
	err = collection.FindOne(ctx, bson.M{
		"$or": []bson.M{
			{"email": req.Email},
			{"username": req.Username},
		},
	}).Decode(&existingUser)

	if err == nil {
		c.JSON(http.StatusConflict, ErrorResponse{Error: "User with this email or username already exists"})
		return
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to hash password"})
		return
	}

	// Create user
	now := time.Now()
	user := models.User{
		ID:          primitive.NewObjectID(),
		Name:        req.Name,
		Email:       req.Email,
		Username:    req.Username,
		Password:    hashedPassword,
		Phone:       req.Phone,
		Department:  req.Department,
		Bio:         req.Bio,
		Role:        req.Role,
		Status:      models.StatusActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	_, err = collection.InsertOne(ctx, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, user.ToResponse())
}

// UpdateUser godoc
// @Summary Update user
// @Description Update an existing user (Admin can update all, Manager can update staff only)
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Param request body models.UpdateUserRequest true "User update data"
// @Success 200 {object} models.UserResponse
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{id} [put]
func UpdateUser(c *gin.Context) {
	id := c.Param("id")
	userObjectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid user ID"})
		return
	}

	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Get current user from context (set by auth middleware)
	currentUserID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	collection := database.DB.Collection("users")
	ctx := context.Background()

	currentUserObjectID, err := primitive.ObjectIDFromHex(currentUserID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid current user ID"})
		return
	}

	var currentUser models.User
	err = collection.FindOne(ctx, bson.M{"_id": currentUserObjectID}).Decode(&currentUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get current user"})
		return
	}

	var user models.User
	err = collection.FindOne(ctx, bson.M{"_id": userObjectID}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "User not found"})
		return
	}

	// Permission checks
	if currentUser.Role == models.RoleAdmin {
		// Admin can update all users except changing other admins' roles
		if req.Role != "" && user.Role == models.RoleAdmin && req.Role != models.RoleAdmin {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: "Admin cannot change other admins' roles"})
			return
		}
	} else if currentUser.Role == models.RoleManager {
		// Manager can only update staff members
		if user.Role != models.RoleStaff {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: "Manager can only update staff accounts"})
			return
		}
		// Manager cannot change roles
		if req.Role != "" {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: "Manager cannot change user roles"})
			return
		}
	} else {
		// Staff cannot update users
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "Staff cannot update user accounts"})
		return
	}

	// Check for email/username conflicts if they're being updated
	if req.Email != "" && req.Email != user.Email {
		var existingUser models.User
		err := collection.FindOne(ctx, bson.M{"email": req.Email, "_id": bson.M{"$ne": userObjectID}}).Decode(&existingUser)
		if err == nil {
			c.JSON(http.StatusConflict, ErrorResponse{Error: "Email already in use"})
			return
		}
		user.Email = req.Email
	}

	if req.Username != "" && req.Username != user.Username {
		var existingUser models.User
		err := collection.FindOne(ctx, bson.M{"username": req.Username, "_id": bson.M{"$ne": userObjectID}}).Decode(&existingUser)
		if err == nil {
			c.JSON(http.StatusConflict, ErrorResponse{Error: "Username already in use"})
			return
		}
		user.Username = req.Username
	}

	// Update other fields
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	if req.Department != "" {
		user.Department = req.Department
	}
	if req.Bio != "" {
		user.Bio = req.Bio
	}
	if req.ProfileImage != "" {
		user.ProfileImage = req.ProfileImage
	}
	if req.SocialLinks != nil {
		user.SocialLinks = req.SocialLinks
	}
	if req.Role != "" {
		user.Role = req.Role
	}
	if req.Status != "" {
		user.Status = req.Status
	}

	user.UpdatedAt = time.Now()

	update := bson.M{"$set": bson.M{
		"name":         user.Name,
		"email":        user.Email,
		"username":     user.Username,
		"phone":        user.Phone,
		"department":   user.Department,
		"bio":          user.Bio,
		"profile_image": user.ProfileImage,
		"social_links": user.SocialLinks,
		"role":         user.Role,
		"status":       user.Status,
		"updated_at":   user.UpdatedAt,
	}}

	_, err = collection.UpdateOne(ctx, bson.M{"_id": userObjectID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, user.ToResponse())
}

// DeleteUser godoc
// @Summary Delete user
// @Description Delete a user account (Admin cannot delete other admins or managers)
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 204 {object} nil
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{id} [delete]
func DeleteUser(c *gin.Context) {
	id := c.Param("id")
	userObjectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid user ID"})
		return
	}

	collection := database.DB.Collection("users")
	ctx := context.Background()

	var user models.User
	err = collection.FindOne(ctx, bson.M{"_id": userObjectID}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "User not found"})
		return
	}

	// Get current user from context (set by auth middleware)
	currentUserID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	currentUserObjectID, err := primitive.ObjectIDFromHex(currentUserID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid current user ID"})
		return
	}

	var currentUser models.User
	err = collection.FindOne(ctx, bson.M{"_id": currentUserObjectID}).Decode(&currentUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get current user"})
		return
	}

	// Admin cannot delete other admins or managers
	if currentUser.Role == models.RoleAdmin && (user.Role == models.RoleAdmin || user.Role == models.RoleManager) {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "Admin cannot delete other admins or managers"})
		return
	}

	// Manager cannot delete admins
	if currentUser.Role == models.RoleManager && user.Role == models.RoleAdmin {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "Manager cannot delete admin"})
		return
	}

	_, err = collection.DeleteOne(ctx, bson.M{"_id": userObjectID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to delete user"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
