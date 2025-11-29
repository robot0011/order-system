package handler

import (
	"order-system/database"
	"order-system/models"

	"github.com/gofiber/fiber/v2"
)

// CreateMenuItem godoc
// @Summary Create a new menu item
// @Description Create a new menu item for a restaurant
// @Tags Menu
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param restaurant_id path string true "Restaurant ID"
// @Param menu_item body MenuItem true "Menu item data"
// @Success 201 {object} MenuItem
// @Failure 400 {string} string "Invalid input"
// @Failure 404 {string} string "Restaurant not found"
// @Failure 500 {string} string "Error creating menu item"
// @Router /api/restaurant/{restaurant_id}/menu [post]
func CreateMenuItem(c *fiber.Ctx) error {
	username := c.Locals("username").(string)
	restaurantID := c.Params("restaurant_id")

	restaurant, err := verifyRestaurantOwnership(username, parseUint(restaurantID))
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Restaurant not found")
	}

	var request struct {
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Price       float64 `json:"price"`
		Category    string  `json:"category"`
		ImageURL    string  `json:"image_url"`
		Quantity    int     `json:"quantity"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid input")
	}

	menuItem := models.MenuItem{
		RestaurantID: restaurant.ID,
		Name:         request.Name,
		Description:  request.Description,
		Price:        request.Price,
		Category:     request.Category,
		ImageURL:     request.ImageURL,
		Quantity:     request.Quantity,
	}

	if err := database.DB.Create(&menuItem).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error creating menu item")
	}

	return c.Status(fiber.StatusCreated).JSON(menuItem)
}

// GetMenuItems godoc
// @Summary Get all menu items
// @Description Get all menu items for a restaurant
// @Tags Menu
// @Produce json
// @Security BearerAuth
// @Param restaurant_id path string true "Restaurant ID"
// @Success 200 {array} MenuItem
// @Failure 404 {string} string "Restaurant not found"
// @Failure 500 {string} string "Error retrieving menu items"
// @Router /api/restaurant/{restaurant_id}/menu [get]
func GetMenuItems(c *fiber.Ctx) error {
	username := c.Locals("username").(string)
	restaurantID := c.Params("restaurant_id")

	restaurant, err := verifyRestaurantOwnership(username, parseUint(restaurantID))
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Restaurant not found")
	}

	var menuItems []models.MenuItem
	if err := database.DB.Where("restaurant_id = ?", restaurant.ID).Find(&menuItems).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error retrieving menu items")
	}

	return c.JSON(menuItems)
}

// UpdateMenuItem godoc
// @Summary Update a menu item
// @Description Update a menu item
// @Tags Menu
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param restaurant_id path string true "Restaurant ID"
// @Param id path string true "Item ID"
// @Param menu_item body MenuItem true "Menu item data"
// @Success 200 {object} MenuItem
// @Failure 400 {string} string "Invalid input"
// @Failure 404 {string} string "Restaurant or menu item not found"
// @Failure 500 {string} string "Error updating menu item"
// @Router /api/restaurant/{restaurant_id}/menu/{id} [put]
func UpdateMenuItem(c *fiber.Ctx) error {
	username := c.Locals("username").(string)
	restaurantID := c.Params("restaurant_id")
	itemID := c.Params("id")

	restaurant, err := verifyRestaurantOwnership(username, parseUint(restaurantID))
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Restaurant not found")
	}

	var menuItem models.MenuItem
	if err := database.DB.Where("id = ? AND restaurant_id = ?", itemID, restaurant.ID).First(&menuItem).Error; err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Menu item not found")
	}

	var request struct {
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Price       float64 `json:"price"`
		Category    string  `json:"category"`
		ImageURL    string  `json:"image_url"`
		Quantity    int     `json:"quantity"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid input")
	}

	menuItem.Name = request.Name
	menuItem.Description = request.Description
	menuItem.Price = request.Price
	menuItem.Category = request.Category
	menuItem.ImageURL = request.ImageURL
	menuItem.Quantity = request.Quantity

	if err := database.DB.Save(&menuItem).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error updating menu item")
	}

	return c.JSON(menuItem)
}

// DeleteMenuItem godoc
// @Summary Delete a menu item
// @Description Delete a menu item
// @Tags Menu
// @Produce json
// @Security BearerAuth
// @Param restaurant_id path string true "Restaurant ID"
// @Param id path string true "Item ID"
// @Success 200 {object} string
// @Failure 404 {string} string "Restaurant or menu item not found"
// @Failure 500 {string} string "Error deleting menu item"
// @Router /api/restaurant/{restaurant_id}/menu/{id} [delete]
func DeleteMenuItem(c *fiber.Ctx) error {
	username := c.Locals("username").(string)
	restaurantID := c.Params("restaurant_id")
	itemID := c.Params("id")

	restaurant, err := verifyRestaurantOwnership(username, parseUint(restaurantID))
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Restaurant not found")
	}

	var menuItem models.MenuItem
	if err := database.DB.Where("id = ? AND restaurant_id = ?", itemID, restaurant.ID).First(&menuItem).Error; err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Menu item not found")
	}

	if err := database.DB.Delete(&menuItem).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error deleting menu item")
	}

	return c.SendStatus(fiber.StatusOK)
}

// GetPublicMenuItems godoc
// @Summary Get public menu items
// @Description Get all menu items for a restaurant without authentication
// @Tags Menu
// @Produce json
// @Param restaurant_id path string true "Restaurant ID"
// @Success 200 {array} MenuItem
// @Failure 404 {string} string "Restaurant not found"
// @Failure 500 {string} string "Error retrieving menu items"
// @Router /api/restaurants/{restaurant_id}/menu [get]
func GetPublicMenuItems(c *fiber.Ctx) error {
	restaurantID := c.Params("restaurant_id")

	// Check if restaurant exists
	var restaurant models.Restaurant
	if err := database.DB.First(&restaurant, restaurantID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Restaurant not found")
	}
	println(restaurant.Name)

	var menuItems []models.MenuItem
	if err := database.DB.Where("restaurant_id = ?", restaurant.ID).Find(&menuItems).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error retrieving menu items")
	}

	return c.JSON(menuItems)
}
