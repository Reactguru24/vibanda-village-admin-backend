package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"gorm.io/gorm"
)

type UserRole string

const (
	RoleAdmin   UserRole = "admin"
	RoleManager UserRole = "manager"
	RoleStaff   UserRole = "staff"
)

type UserStatus string

const (
	StatusActive   UserStatus = "active"
	StatusInactive UserStatus = "inactive"
)

// User represents a user in the system
type User struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty" gorm:"type:objectid;primaryKey;autoIncrement:false"`
	Name        string             `json:"name" bson:"name" gorm:"not null" validate:"required,min=2,max=100"`
	Email       string             `json:"email" bson:"email" gorm:"uniqueIndex;not null" validate:"required,email"`
	Username    string             `json:"username" bson:"username" gorm:"uniqueIndex;not null" validate:"required,min=3,max=50"`
	Password    string             `json:"-" bson:"password" gorm:"not null" validate:"required,min=6"`
	Role        UserRole           `json:"role" bson:"role" gorm:"not null" validate:"required,oneof=admin manager staff"`
	Status      UserStatus         `json:"status" bson:"status" gorm:"not null;default:active" validate:"required,oneof=active inactive"`
	Phone       string             `json:"phone,omitempty" bson:"phone,omitempty"`
	Department  string             `json:"department,omitempty" bson:"department,omitempty"`
	Bio         string             `json:"bio,omitempty" bson:"bio,omitempty"`
	ProfileImage string            `json:"profile_image,omitempty" bson:"profile_image,omitempty"`
	SocialLinks map[string]string  `json:"social_links,omitempty" bson:"social_links,omitempty"`
	LastLogin   *time.Time         `json:"last_login,omitempty" bson:"last_login,omitempty"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
}

// BeforeCreate hook to set ID and timestamps
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID.IsZero() {
		u.ID = primitive.NewObjectID()
	}
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
	return nil
}

// BeforeUpdate hook to update timestamp
func (u *User) BeforeUpdate(tx *gorm.DB) error {
	u.UpdatedAt = time.Now()
	return nil
}

// UserResponse represents user data returned to client (without password)
type UserResponse struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Email       string            `json:"email"`
	Username    string            `json:"username"`
	Role        UserRole          `json:"role"`
	Status      UserStatus        `json:"status"`
	Phone       string            `json:"phone,omitempty"`
	Department  string            `json:"department,omitempty"`
	Bio         string            `json:"bio,omitempty"`
	ProfileImage string           `json:"profile_image,omitempty"`
	SocialLinks map[string]string `json:"social_links,omitempty"`
	LastLogin   *time.Time        `json:"last_login,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// ToResponse converts User to UserResponse
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:          u.ID.Hex(),
		Name:        u.Name,
		Email:       u.Email,
		Username:    u.Username,
		Role:        u.Role,
		Status:      u.Status,
		Phone:       u.Phone,
		Department:  u.Department,
		Bio:         u.Bio,
		ProfileImage: u.ProfileImage,
		SocialLinks: u.SocialLinks,
		LastLogin:   u.LastLogin,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}
}

// LoginRequest represents login request payload
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// RegisterRequest represents registration request payload
type RegisterRequest struct {
	Name       string   `json:"name" validate:"required,min=2,max=100"`
	Email      string   `json:"email" validate:"required,email"`
	Username   string   `json:"username" validate:"required,min=3,max=50"`
	Password   string   `json:"password" validate:"required,min=6"`
	Phone      string   `json:"phone,omitempty"`
	Department string   `json:"department,omitempty"`
	Bio        string   `json:"bio,omitempty"`
	Role       UserRole `json:"role" validate:"required,oneof=admin manager staff"`
}

// UpdateUserRequest represents user update request payload
type UpdateUserRequest struct {
	Name        string            `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Email       string            `json:"email,omitempty" validate:"omitempty,email"`
	Username    string            `json:"username,omitempty" validate:"omitempty,min=3,max=50"`
	Phone       string            `json:"phone,omitempty"`
	Department  string            `json:"department,omitempty"`
	Bio         string            `json:"bio,omitempty"`
	ProfileImage string           `json:"profile_image,omitempty"`
	SocialLinks map[string]string `json:"social_links,omitempty"`
	Role        UserRole          `json:"role,omitempty" validate:"omitempty,oneof=admin manager staff"`
	Status      UserStatus        `json:"status,omitempty" validate:"omitempty,oneof=active inactive"`
}

// ProfileActivity represents a user activity entry
type ProfileActivity struct {
	ID        string `json:"id"`
	Description string `json:"description"`
	Timestamp time.Time `json:"timestamp"`
}

// ProfilePermissions represents role-based permissions
type ProfilePermissions struct {
	CanManageUsers     bool     `json:"can_manage_users"`
	CanManageRoles     bool     `json:"can_manage_roles"`
	CanManageSystem    bool     `json:"can_manage_system"`
	AccessPermissions  []string `json:"access_permissions"`
}

// ProfileResponse represents comprehensive profile data for UserProfile.vue
type ProfileResponse struct {
	// Basic user info
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Email       string            `json:"email"`
	Username    string            `json:"username"`
	Role        UserRole          `json:"role"`
	Status      UserStatus        `json:"status"`
	Phone       string            `json:"phone,omitempty"`
	Department  string            `json:"department,omitempty"`
	Bio         string            `json:"bio,omitempty"`
	ProfileImage string           `json:"profile_image,omitempty"`
	SocialLinks map[string]string `json:"social_links,omitempty"`
	LastLogin   *time.Time        `json:"last_login,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`

	// Profile-specific data
	JoinDate         string             `json:"join_date"`
	RoleDisplay      string             `json:"role_display"`
	Permissions      ProfilePermissions `json:"permissions"`
	RecentActivities []ProfileActivity  `json:"recent_activities"`
}
