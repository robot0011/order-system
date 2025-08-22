package main

import (
	"order-system/handler"

	"github.com/gofiber/fiber/v2"
)

func setupRoutes(app *fiber.App) {
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"message": "Welcome to the Notes API",
		})
	})
	api := app.Group("/api")

	user := api.Group("/user")
	user.Get("/", handler.GetAllUsers) // New route to get all users
	user.Post("/register", handler.Register)
	user.Post("/login", handler.Login)
	user.Get("/profile", handler.ProtectRoute, handler.Profile)
	user.Post("/refresh", handler.RefreshToken)

	tenant := api.Group("/tenant")
	tenant.Post("/create", handler.ProtectRoute, handler.CreateTenant)
	tenant.Get("/:tenant_id/:role_id", handler.ProtectRoute, handler.SelectTenantAndRole)
	tenant.Post("/:tenant_id/menu-items", handler.ProtectRoute, handler.CreateMenuItem)
	tenant.Get("/:tenant_id/menu-items", handler.ProtectRoute, handler.GetMenuItems)
	tenant.Delete("/:tenant_id/menu-items/:item_id", handler.ProtectRoute, handler.DeleteMenuItem)
	tenant.Patch("/:tenant_id/menu-items/:item_id", handler.ProtectRoute, handler.UpdateMenuItem)

}
