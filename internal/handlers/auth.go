package handlers

import (
	"context"
	"net/http"
	"time"
	"vibanda-village-backend/internal/config"
	"vibanda-village-backend/internal/database"
	"vibanda-village-backend/internal/models"
	"vibanda-village-backend/pkg/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// LoginResponse represents login response
type LoginResponse struct {
	Token string              `json:"token"`
	User  models.UserResponse `json:"user"`
}

// Register godoc
// @Summary Register a new user
// @Description Register a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.RegisterRequest true "Registration data"
// @Success 201 {object} models.UserResponse
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/register [post]
func Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Check if user already exists
	collection := database.DB.Collection("users")
	ctx := context.Background()

	var existingUser models.User
	err := collection.FindOne(ctx, bson.M{
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
		ID:        primitive.NewObjectID(),
		Name:      req.Name,
		Email:     req.Email,
		Username:  req.Username,
		Password:  hashedPassword,
		Phone:     req.Phone,
		Role:      req.Role,
		Status:    models.StatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}

	_, err = collection.InsertOne(ctx, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, user.ToResponse())
}

// Login godoc
// @Summary Login user
// @Description Authenticate user and return JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "Login credentials"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/login [post]
func Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Find user by email
	collection := database.DB.Collection("users")
	ctx := context.Background()

	var user models.User
	err := collection.FindOne(ctx, bson.M{"email": req.Email}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid credentials"})
		return
	}

	// Check password
	if !utils.CheckPassword(req.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid credentials"})
		return
	}

	// Check if user is active
	if user.Status != models.StatusActive {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Account is not active"})
		return
	}

	// Update last login
	now := time.Now()
	user.LastLogin = &now
	user.UpdatedAt = now

	update := bson.M{"$set": bson.M{"last_login": user.LastLogin, "updated_at": user.UpdatedAt}}
	_, err = collection.UpdateOne(ctx, bson.M{"_id": user.ID}, update)
	if err != nil {
		// Log error but don't fail login
	}

	// Generate JWT token
	cfg := config.Load()
	token, err := utils.GenerateToken(&user, cfg.JWTSecret, cfg.JWTExpirationHours)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to generate token"})
		return
	}

	response := LoginResponse{
		Token: token,
		User:  user.ToResponse(),
	}

	c.JSON(http.StatusOK, response)
}

// GetProfile godoc
// @Summary Get user profile
// @Description Get current user profile information
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.UserResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/profile [get]
func GetProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")
	userIDStr := userID.(string)

	userObjectID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid user ID"})
		return
	}

	collection := database.DB.Collection("users")
	ctx := context.Background()

	var user models.User
	err = collection.FindOne(ctx, bson.M{"_id": userObjectID}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get user profile"})
		return
	}

	c.JSON(http.StatusOK, user.ToResponse())
}
