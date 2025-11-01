package main

import (
	"context"
	"log"
	"time"
	"vibanda-village-admin-backend/internal/config"
	"vibanda-village-admin-backend/internal/database"
	"vibanda-village-admin-backend/internal/models"
	"vibanda-village-admin-backend/pkg/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func main() {
	cfg := config.Load()
	database.InitDB(cfg.MongoURI, cfg.DatabaseName)

	collection := database.DB.Collection("users")
	ctx := context.Background()

	// Check if user already exists
	var existingUser models.User
	err := collection.FindOne(ctx, bson.M{"email": "testandtest@gmail.com"}).Decode(&existingUser)
	if err == nil {
		log.Println("User already exists")
		return
	}

	// Hash password
	hashedPassword, err := utils.HashPassword("12345678")
	if err != nil {
		log.Fatal("Failed to hash password:", err)
	}

	// Create user
	now := time.Now()
	user := models.User{
		ID:        primitive.NewObjectID(),
		Name:      "Test User",
		Email:     "testandtest@gmail.com",
		Username:  "testuser",
		Password:  hashedPassword,
		Phone:     "",
		Role:      models.RoleAdmin,
		Status:    models.StatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}

	_, err = collection.InsertOne(ctx, user)
	if err != nil {
		log.Fatal("Failed to create user:", err)
	}

	log.Println("Test user created successfully")
}
