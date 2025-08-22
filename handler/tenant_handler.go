package handler

import (
	"fmt"
	"order-system/database"
	"order-system/models"

	"github.com/gofiber/fiber/v2"
)

func CreateTenant(c *fiber.Ctx) error {
	var input struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Location    string `json:"location"`
		Logo        string `json:"logo"`
	}

	// Parse the request body
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	// Get the user who is creating the tenant
	username := c.Locals("username")
	var user models.User
	database.DB.Where("username = ?", username).First(&user)

	// Create the tenant
	tenant := models.Tenant{
		Name:        input.Name,
		Description: input.Description,
		Location:    input.Location,
		Logo:        input.Logo,
		Status:      "Active", // Default status is "Active"
		OwnerID:     user.ID,  // Set the owner of the tenant
	}

	// Save the tenant in the database
	if err := database.DB.Create(&tenant).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error creating tenant")
	}

	// Return the created tenant details
	return c.JSON(fiber.Map{
		"message": "Tenant created successfully",
		"tenant":  tenant,
	})
}

func GetTenants(c *fiber.Ctx) error {
	var tenants []models.Tenant
	err := database.DB.Find(&tenants).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Could not retrieve tenants")
	}

	return c.JSON(tenants)
}

func SelectTenantAndRole(c *fiber.Ctx) error {
	tenantID := c.Params("tenant_id")
	roleID := c.Params("role_id") // Role selected by the user

	var tenant models.Tenant
	if err := database.DB.Where("id = ?", tenantID).First(&tenant).Error; err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Tenant not found")
	}

	// Fetch tasks for the selected role in the selected tenant
	var tasks []models.Task
	database.DB.Where("role_id = ?", roleID).Find(&tasks)

	userID := c.Locals("user_id").(uint)

	// Convert tenantID and roleID from string to uint
	var tenantIDUint uint
	var roleIDUint uint
	if _, err := fmt.Sscan(tenantID, &tenantIDUint); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid tenant_id")
	}
	if _, err := fmt.Sscan(roleID, &roleIDUint); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid role_id")
	}

	userTenantRole := models.UserTenantRole{
		UserID:   userID,
		TenantID: tenantIDUint,
		RoleID:   roleIDUint,
		Active:   true,
	}
	database.DB.Create(&userTenantRole)

	return c.JSON(fiber.Map{
		"message": "Role selected",
		"tasks":   tasks, // Display tasks associated with the selected role
	})
}

func CreateMenuItem(c *fiber.Ctx) error {
	var input struct {
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Price       float64 `json:"price"`
		Category    string  `json:"category"`
		TenantID    uint    `json:"tenant_id"` // The tenant adding the food item
	}

	// Parse the request body
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	// Ensure the tenant exists (optional)
	var tenant models.Tenant
	if err := database.DB.Where("id = ?", input.TenantID).First(&tenant).Error; err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Tenant not found")
	}

	// Create a new menu item
	menuItem := models.MenuItem{
		Name:        input.Name,
		Description: input.Description,
		Price:       input.Price,
		Category:    input.Category,
		TenantID:    input.TenantID,
	}

	// Save the menu item to the database
	if err := database.DB.Create(&menuItem).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error creating menu item")
	}

	return c.JSON(fiber.Map{
		"message": "Menu item added successfully",
		"item":    menuItem,
	})
}

func UpdateMenuItem(c *fiber.Ctx) error {
	itemID := c.Params("item_id")

	// Find the menu item by ID
	var menuItem models.MenuItem
	if err := database.DB.Where("id = ?", itemID).First(&menuItem).Error; err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Menu item not found")
	}

	var input struct {
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Price       float64 `json:"price"`
		Category    string  `json:"category"`
	}

	// Parse the request body to update the item
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	// Update menu item fields
	menuItem.Name = input.Name
	menuItem.Description = input.Description
	menuItem.Price = input.Price
	menuItem.Category = input.Category

	// Save the changes to the database
	if err := database.DB.Save(&menuItem).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error updating menu item")
	}

	return c.JSON(fiber.Map{
		"message": "Menu item updated successfully",
		"item":    menuItem,
	})
}

func DeleteMenuItem(c *fiber.Ctx) error {
	itemID := c.Params("item_id")

	// Find the menu item by ID
	if err := database.DB.Where("id = ?", itemID).Delete(&models.MenuItem{}).Error; err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Menu item not found")
	}

	return c.SendStatus(fiber.StatusOK)
}

func GetMenuItems(c *fiber.Ctx) error {
	tenantID := c.Params("tenant_id")

	// Fetch all menu items for the specified tenant
	var menuItems []models.MenuItem
	if err := database.DB.Where("tenant_id = ?", tenantID).Find(&menuItems).Error; err != nil {
		return c.Status(fiber.StatusNotFound).SendString("No menu items found for this tenant")
	}

	return c.JSON(fiber.Map{
		"message":    "Menu items fetched successfully",
		"menu_items": menuItems,
	})
}
