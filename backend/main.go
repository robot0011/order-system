package main

import (
	"log"
	"order-system/database"
	_ "order-system/docs"
	"os"
	"strings"

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

	// CORS configuration - security: cannot use wildcard with credentials
	corsOrigins := os.Getenv("CORS_ORIGINS")
	allowOrigins := []string{}
	if corsOrigins == "" {
		// In development, you can set specific origins
		// For production, always define specific origins
		allowOrigins = []string{
			"http://localhost:5173", // Vite default port
			"http://localhost:3000", // Common React port
			"http://localhost:3001", // Alternative port
		}
	} else {
		// Split comma-separated origins
		origins := strings.Split(corsOrigins, ",")
		for _, origin := range origins {
			allowOrigins = append(allowOrigins, strings.TrimSpace(origin))
		}
	}

	// Add CORS middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins:     strings.Join(allowOrigins, ","),
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
