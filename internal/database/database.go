package database

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"vibanda-village-admin-backend/internal/models"
	"vibanda-village-admin-backend/pkg/utils"
)

var Client *mongo.Client
var DB *mongo.Database

func InitDB(mongoURI, databaseName string) {
	// Connect to MongoDB using mongo-driver with retry options
	clientOptions := options.Client().
		ApplyURI(mongoURI).
		SetServerSelectionTimeout(5 * time.Second).
		SetConnectTimeout(10 * time.Second).
		SetMaxPoolSize(10)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}

	// Ping the database with retry
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)

		if err := client.Ping(pingCtx, nil); err != nil {
			pingCancel()
			if i == maxRetries-1 {
				log.Fatal("Failed to ping MongoDB after", maxRetries, "attempts:", err)
			}
			log.Printf("MongoDB ping attempt %d failed, retrying...", i+1)
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}

		pingCancel()
		break
	}

	log.Println("Connected to MongoDB successfully")

	Client = client
	DB = client.Database(databaseName)
	log.Println("Database connection established")

	// Create test user if it doesn't exist
	createTestUserIfNotExists()
}

func GetClient() *mongo.Client {
	return Client
}

func GetDatabase() *mongo.Database {
	return DB
}

func CloseDB() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := Client.Disconnect(ctx); err != nil {
		log.Println("Error closing database connection:", err)
	}
}

func createTestUserIfNotExists() {
	collection := DB.Collection("users")
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
		log.Println("Failed to hash password:", err)
		return
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
		log.Println("Failed to create user:", err)
		return
	}

	log.Println("Test user created successfully")
}
