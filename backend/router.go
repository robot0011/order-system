package main

import (
	"net/http"
	"order-system/handler"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

func healthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":  "ok",
		"message": "Welcome to the Order-System API",
	})
}

func setupRoutes(app *fiber.App) {
	app.Get("/health", handler.HealthCheck)
	app.Get("/ws/orders", websocket.New(handler.HandleOrderSocket, websocket.Config{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}))
	api := app.Group("/api")

	user := api.Group("/user")
	user.Get("/", handler.GetAllUsers)
	user.Post("/register", handler.Register)
	user.Post("/login", handler.Login)
	user.Get("/profile", handler.ProtectRoute, handler.Profile)
	user.Post("/refresh", handler.RefreshToken)
	user.Delete("/", handler.DeleteUser)

	// Public restaurant endpoints (no authentication required)
	api.Get("/restaurant/:id", handler.GetPublicRestaurantByID)
	api.Get("/restaurants/:restaurant_id/menu", handler.GetPublicMenuItems)  // Different route to avoid conflict
	api.Post("/restaurants/:restaurant_id/order", handler.CreatePublicOrder) // Different route to avoid conflict

	// Protected restaurant management endpoints (authentication required)
	protectedRestaurant := api.Group("/restaurant", handler.ProtectRoute)
	protectedRestaurant.Post("/", handler.CreateRestaurant)
	protectedRestaurant.Get("/", handler.GetRestaurants)
	protectedRestaurant.Get("/:id", handler.GetRestaurantByID) // Allow authenticated users to get restaurant details too
	protectedRestaurant.Put("/:id", handler.UpdateRestaurant)
	protectedRestaurant.Delete("/:id", handler.DeleteRestaurant)

	// Table routes (nested under restaurant - protected)
	protectedRestaurant.Post("/:restaurant_id/table", handler.CreateTable)
	protectedRestaurant.Get("/:restaurant_id/table", handler.GetTables)
	protectedRestaurant.Put("/:restaurant_id/table/:id", handler.UpdateTable)
	protectedRestaurant.Delete("/:restaurant_id/table/:id", handler.DeleteTable)

	// All tables route (for all restaurants the user owns)
	api.Get("/table", handler.ProtectRoute, handler.GetAllUserTables)

	// Menu routes (nested under restaurant - protected)
	protectedRestaurant.Post("/:restaurant_id/menu", handler.CreateMenuItem)
	protectedRestaurant.Get("/:restaurant_id/menu", handler.GetMenuItems) // Protected access to owner's menu
	protectedRestaurant.Put("/:restaurant_id/menu/:id", handler.UpdateMenuItem)
	protectedRestaurant.Delete("/:restaurant_id/menu/:id", handler.DeleteMenuItem)

	// Order routes (nested under restaurant - protected)
	protectedRestaurant.Get("/:restaurant_id/order", handler.GetOrders)
	protectedRestaurant.Get("/:restaurant_id/order/:id", handler.GetOrder)
	protectedRestaurant.Patch("/:restaurant_id/order/:id", handler.UpdateOrderStatus)
	protectedRestaurant.Delete("/:restaurant_id/order/:id", handler.DeleteOrder)

	// All orders route (for all restaurants the user owns)
	api.Get("/order", handler.ProtectRoute, handler.GetAllUserOrders)
}
