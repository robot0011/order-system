package handler

import (
	"order-system/database"
	"order-system/models"

	"github.com/gofiber/fiber/v2"
)

// CreateRestaurant godoc
// @Summary Create a new restaurant
// @Description Create a new restaurant for the authenticated user
// @Tags Restaurant
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param restaurant body Restaurant true "Restaurant data"
// @Success 201 {object} Restaurant
// @Failure 400 {string} string "Invalid input"
// @Failure 404 {string} string "User not found"
// @Failure 500 {string} string "Error creating restaurant"
// @Router /api/restaurant/ [post]
func CreateRestaurant(c *fiber.Ctx) error {
	username := c.Locals("username").(string)

	var user models.User
	if err := database.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).SendString("User not found")
	}

	var request struct {
		Name        string `json:"name"`
		Address     string `json:"address"`
		PhoneNumber string `json:"phone_number"`
		LogoURL     string `json:"logo_url"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid input")
	}

	restaurant := models.Restaurant{
		UserID:      user.ID,
		Name:        request.Name,
		Address:     request.Address,
		PhoneNumber: request.PhoneNumber,
		LogoURL:     request.LogoURL,
	}

	if err := database.DB.Create(&restaurant).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error creating restaurant")
	}

	return c.Status(fiber.StatusCreated).JSON(restaurant)
}

// GetRestaurants godoc
// @Summary Get all restaurants
// @Description Get all restaurants for the authenticated user
// @Tags Restaurant
// @Produce json
// @Security BearerAuth
// @Success 200 {array} Restaurant
// @Failure 404 {string} string "User not found"
// @Failure 500 {string} string "Error retrieving restaurants"
// @Router /api/restaurant/ [get]
func GetRestaurants(c *fiber.Ctx) error {
	username := c.Locals("username").(string)

	var user models.User
	if err := database.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).SendString("User not found")
	}

	var restaurants []models.Restaurant
	if err := database.DB.Where("user_id = ?", user.ID).Find(&restaurants).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error retrieving restaurants")
	}

	return c.JSON(restaurants)
}

// GetRestaurantByID godoc
// @Summary Get restaurant by ID
// @Description Get a restaurant by ID
// @Tags Restaurant
// @Produce json
// @Security BearerAuth
// @Param id path string true "Restaurant ID"
// @Success 200 {object} Restaurant
// @Failure 404 {string} string "User or restaurant not found"
// @Failure 500 {string} string "Error retrieving restaurant"
// @Router /api/restaurant/{id} [get]
func GetRestaurantByID(c *fiber.Ctx) error {
	username := c.Locals("username").(string)
	id := c.Params("id")

	var user models.User
	if err := database.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).SendString("User not found")
	}

	var restaurant models.Restaurant
	if err := database.DB.Where("id = ? AND user_id = ?", id, user.ID).Preload("Tables").Preload("MenuItems").First(&restaurant).Error; err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Restaurant not found")
	}

	return c.JSON(restaurant)
}

// UpdateRestaurant godoc
// @Summary Update restaurant
// @Description Update a restaurant by ID
// @Tags Restaurant
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Restaurant ID"
// @Param restaurant body Restaurant true "Restaurant data"
// @Success 200 {object} Restaurant
// @Failure 400 {string} string "Invalid input"
// @Failure 404 {string} string "User or restaurant not found"
// @Failure 500 {string} string "Error updating restaurant"
// @Router /api/restaurant/{id} [put]
func UpdateRestaurant(c *fiber.Ctx) error {
	username := c.Locals("username").(string)
	id := c.Params("id")

	var user models.User
	if err := database.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).SendString("User not found")
	}

	var restaurant models.Restaurant
	if err := database.DB.Where("id = ? AND user_id = ?", id, user.ID).First(&restaurant).Error; err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Restaurant not found")
	}

	var request struct {
		Name        string `json:"name"`
		Address     string `json:"address"`
		PhoneNumber string `json:"phone_number"`
		LogoURL     string `json:"logo_url"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid input")
	}

	restaurant.Name = request.Name
	restaurant.Address = request.Address
	restaurant.PhoneNumber = request.PhoneNumber
	restaurant.LogoURL = request.LogoURL

	if err := database.DB.Save(&restaurant).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error updating restaurant")
	}

	return c.JSON(restaurant)
}

// GetPublicRestaurantByID godoc
// @Summary Get public restaurant by ID
// @Description Get a restaurant by ID without authentication
// @Tags Restaurant
// @Produce json
// @Param id path string true "Restaurant ID"
// @Success 200 {object} Restaurant
// @Failure 404 {string} string "Restaurant not found"
// @Router /api/restaurant/{id} [get]
func GetPublicRestaurantByID(c *fiber.Ctx) error {
	id := c.Params("id")

	var restaurant models.Restaurant
	if err := database.DB.Where("id = ?", id).Preload("Tables").First(&restaurant).Error; err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Restaurant not found")
	}

	return c.JSON(restaurant)
}

// DeleteRestaurant godoc
// @Summary Delete restaurant
// @Description Delete a restaurant by ID
// @Tags Restaurant
// @Produce json
// @Security BearerAuth
// @Param id path string true "Restaurant ID"
// @Success 200 {object} string
// @Failure 404 {string} string "User or restaurant not found"
// @Failure 500 {string} string "Error deleting restaurant"
// @Router /api/restaurant/{id} [delete]
func DeleteRestaurant(c *fiber.Ctx) error {
	username := c.Locals("username").(string)
	id := c.Params("id")

	var user models.User
	if err := database.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).SendString("User not found")
	}

	var restaurant models.Restaurant
	if err := database.DB.Where("id = ? AND user_id = ?", id, user.ID).First(&restaurant).Error; err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Restaurant not found")
	}

	if err := database.DB.Delete(&restaurant).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error deleting restaurant")
	}

	return c.SendStatus(fiber.StatusOK)
}
