package main

import (
	"order-system/database"

	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	database.ConnectDB()
	app := fiber.New()

	// CORS configuration
	corsOrigins := os.Getenv("CORS_ORIGINS")
	if corsOrigins == "" {
		corsOrigins = "http://localhost:5173,http://localhost:3000,http://localhost:8080"
	}

	// Add CORS middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins:     corsOrigins,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
		AllowCredentials: true,
		ExposeHeaders:    "Content-Length",
		MaxAge:           86400, // 24 hours
	}))

	app.Use(logger.New())
	setupRoutes(app)
	app.Listen(":" + port)
}
