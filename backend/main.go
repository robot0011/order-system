package main

import (
	"log"
	"order-system/database"
	_ "order-system/docs"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/swagger"
	"github.com/joho/godotenv"
)

// Load environment variables
func init() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}
}

// @title Order System API
// @version 1.0
// @description API for Order System with user authentication and restaurant management
// @host localhost:3000
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

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
		corsOrigins = "*"
	}

	// Add CORS middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins:     corsOrigins,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS, PATCH",
		AllowCredentials: true, // Enable credentials for WebSocket auth
		ExposeHeaders:    "Content-Length",
		MaxAge:           86400, // 24 hours
	}))

	app.Use(logger.New())

	// Swagger route
	app.Get("/swagger/*", swagger.HandlerDefault)

	setupRoutes(app)
	app.Listen(":" + port)
}
