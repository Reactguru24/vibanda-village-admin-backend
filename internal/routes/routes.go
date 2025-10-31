package routes

import (
	"vibanda-village-admin-backend/internal/config"
	"vibanda-village-admin-backend/internal/handlers"
	"vibanda-village-admin-backend/internal/middleware"
	"vibanda-village-admin-backend/internal/models"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRoutes(r *gin.Engine) {
	cfg := config.Load()

	// CORS middleware
	r.Use(middleware.CORSMiddleware(cfg.AllowedOrigins))

	// Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Public routes (no authentication required)
	public := r.Group("/api/v1")
	{
		auth := public.Group("/auth")
		{
			auth.POST("/register", handlers.Register)
			auth.POST("/login", handlers.Login)
		}
	}

	// Protected routes (authentication required)
	protected := r.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware())
	{
		// Auth routes
		auth := protected.Group("/auth")
		{
			auth.GET("/profile", handlers.GetProfile)
		}

		// User management routes (admin only)
		users := protected.Group("/users")
		users.Use(middleware.RoleMiddleware(models.RoleAdmin))
		{
			users.GET("", handlers.GetUsers)
			users.GET("/:id", handlers.GetUser)
			users.POST("", handlers.CreateUser)
			users.PUT("/:id", handlers.UpdateUser)
			users.DELETE("/:id", handlers.DeleteUser)
		}

		// Product routes (admin and manager)
		products := protected.Group("/products")
		products.Use(middleware.RoleMiddleware(models.RoleAdmin, models.RoleManager))
		{
			products.GET("", handlers.GetProducts)
			products.GET("/:id", handlers.GetProduct)
			products.POST("", handlers.CreateProduct)
			products.PUT("/:id", handlers.UpdateProduct)
			products.DELETE("/:id", handlers.DeleteProduct)
		}

		// Order routes (admin and manager)
		orders := protected.Group("/orders")
		orders.Use(middleware.RoleMiddleware(models.RoleAdmin, models.RoleManager))
		{
			orders.GET("", handlers.GetOrders)
			orders.GET("/:id", handlers.GetOrder)
			orders.POST("", handlers.CreateOrder)
			orders.PUT("/:id", handlers.UpdateOrder)
			orders.DELETE("/:id", handlers.DeleteOrder)
		}

		// Event routes (admin and manager)
		events := protected.Group("/events")
		events.Use(middleware.RoleMiddleware(models.RoleAdmin, models.RoleManager))
		{
			events.GET("", handlers.GetEvents)
			events.GET("/:id", handlers.GetEvent)
			events.POST("", handlers.CreateEvent)
			events.PUT("/:id", handlers.UpdateEvent)
			events.DELETE("/:id", handlers.DeleteEvent)
		}

		// Reservation routes (admin and manager)
		reservations := protected.Group("/reservations")
		reservations.Use(middleware.RoleMiddleware(models.RoleAdmin, models.RoleManager))
		{
			reservations.GET("", handlers.GetReservations)
			reservations.GET("/:id", handlers.GetReservation)
			reservations.POST("", handlers.CreateReservation)
			reservations.PUT("/:id", handlers.UpdateReservation)
			reservations.DELETE("/:id", handlers.DeleteReservation)
		}
	}
}
