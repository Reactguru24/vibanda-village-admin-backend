package main

import (
	"log"
	"vibanda-village-backend/internal/config"
	"vibanda-village-backend/internal/database"
	"vibanda-village-backend/internal/routes"

	_ "vibanda-village-backend/docs" // Import generated docs

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// @title Vibanda Village Admin API
// @version 1.0
// @description A comprehensive backend API for Vibanda Village restaurant management system
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@vibandavillage.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize configuration
	cfg := config.Load()

	// Set Gin mode
	gin.SetMode(cfg.GinMode)

	// Initialize database
	database.InitDB(cfg.MongoURI, cfg.DatabaseName)

	// Create Gin router
	r := gin.Default()

	// Setup routes
	routes.SetupRoutes(r)

	// Start server
	serverAddr := ":" + cfg.Port
	log.Printf("🚀 Server starting on http://localhost%s", serverAddr)
	log.Printf("📚 Swagger documentation available at: http://localhost%s/swagger/index.html", serverAddr)
	log.Printf("🔗 API base URL: http://localhost%s/api/v1", serverAddr)

	if err := r.Run(serverAddr); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
