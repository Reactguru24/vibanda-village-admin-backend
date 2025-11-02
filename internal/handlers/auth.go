package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"
	"vibanda-village-admin-backend/internal/config"
	"vibanda-village-admin-backend/internal/database"
	"vibanda-village-admin-backend/internal/models"
	"vibanda-village-admin-backend/pkg/utils"

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
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request format. Please check your input data and try again."})
		return
	}

	fmt.Printf("Register payload: %+v\n", req)

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
		c.JSON(http.StatusConflict, ErrorResponse{Error: "An account with this email or username already exists. Please use a different email or try logging in if you already have an account."})
		return
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "An error occurred while processing your request. Please try again later."})
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
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "An error occurred while creating your account. Please try again later."})
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
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request format. Please check your input data and try again."})
		return
	}

	fmt.Printf("Login payload: %+v\n", req)

	// Find user by email
	collection := database.DB.Collection("users")
	ctx := context.Background()

	var user models.User
	err := collection.FindOne(ctx, bson.M{"email": req.Email}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "The email or password you entered is incorrect. Please check your credentials and try again."})
		return
	}

	// Check password
	if !utils.CheckPassword(req.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "The email or password you entered is incorrect. Please check your credentials and try again."})
		return
	}

	// Check if user is active
	if user.Status != models.StatusActive {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Your account is currently inactive. Please contact support for assistance."})
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
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "An error occurred while logging you in. Please try again later."})
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
// @Description Get current user profile information with role-based data
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.ProfileResponse
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
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "An error occurred while retrieving your profile. Please try again later."})
		return
	}

	// Build role-based permissions
	permissions := getRolePermissions(user.Role)

	// Build role display name
	roleDisplay := getRoleDisplay(user.Role)

	// Build recent activities (mock data for now - in real app, this would come from activity logs)
	recentActivities := []models.ProfileActivity{
		{
			ID:          "1",
			Description: "Logged into admin dashboard",
			Timestamp:   time.Now().Add(-time.Hour * 2),
		},
		{
			ID:          "2",
			Description: "Updated profile information",
			Timestamp:   time.Now().Add(-time.Hour * 24),
		},
		{
			ID:          "3",
			Description: "Created new user account",
			Timestamp:   time.Now().Add(-time.Hour * 48),
		},
	}

	// Create comprehensive profile response
	profileResponse := models.ProfileResponse{
		ID:               user.ID.Hex(),
		Name:             user.Name,
		Email:            user.Email,
		Username:         user.Username,
		Role:             user.Role,
		Status:           user.Status,
		Phone:            user.Phone,
		Department:       user.Department,
		Bio:              user.Bio,
		ProfileImage:     user.ProfileImage,
		SocialLinks:      user.SocialLinks,
		LastLogin:        user.LastLogin,
		CreatedAt:        user.CreatedAt,
		UpdatedAt:        user.UpdatedAt,
		JoinDate:         user.CreatedAt.Format("2006-01-02"),
		RoleDisplay:      roleDisplay,
		Permissions:      permissions,
		RecentActivities: recentActivities,
	}

	c.JSON(http.StatusOK, profileResponse)
}

// Helper function to get role-based permissions
func getRolePermissions(role models.UserRole) models.ProfilePermissions {
	switch role {
	case models.RoleAdmin:
		return models.ProfilePermissions{
			CanManageUsers:  true,
			CanManageRoles:  true,
			CanManageSystem: true,
			AccessPermissions: []string{
				"Full system access",
				"User management",
				"Role assignment",
				"System configuration",
				"Financial reports",
				"Inventory management",
				"Order processing",
				"Reservation management",
				"Event management",
				"Customer data access",
			},
		}
	case models.RoleManager:
		return models.ProfilePermissions{
			CanManageUsers:  true,
			CanManageRoles:  false,
			CanManageSystem: false,
			AccessPermissions: []string{
				"Dashboard access",
				"Team management",
				"Order processing",
				"Reservation management",
				"Event management",
				"Inventory oversight",
				"Staff scheduling",
				"Basic reporting",
			},
		}
	case models.RoleStaff:
		return models.ProfilePermissions{
			CanManageUsers:  false,
			CanManageRoles:  false,
			CanManageSystem: false,
			AccessPermissions: []string{
				"Dashboard access",
				"Order processing",
				"Reservation management",
				"Event assistance",
				"Inventory updates",
				"Customer service",
			},
		}
	default:
		return models.ProfilePermissions{
			CanManageUsers:    false,
			CanManageRoles:    false,
			CanManageSystem:   false,
			AccessPermissions: []string{},
		}
	}
}

// Helper function to get role display name
func getRoleDisplay(role models.UserRole) string {
	switch role {
	case models.RoleAdmin:
		return "System Administrator"
	case models.RoleManager:
		return "Management Team"
	case models.RoleStaff:
		return "Staff Member"
	default:
		return string(role)
	}
}
