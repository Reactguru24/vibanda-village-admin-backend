package database

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
