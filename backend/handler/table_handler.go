package handler

import (
	"fmt"
	"order-system/database"
	"order-system/models"
	"order-system/utils"

	"github.com/gofiber/fiber/v2"
)

// verifyRestaurantOwnership checks if the restaurant belongs to the user
func verifyRestaurantOwnership(username string, restaurantID uint) (*models.Restaurant, error) {
	var user models.User
	if err := database.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}

	var restaurant models.Restaurant
	if err := database.DB.Where("id = ? AND user_id = ?", restaurantID, user.ID).First(&restaurant).Error; err != nil {
		return nil, err
	}

	return &restaurant, nil
}

// CreateTable godoc
// @Summary Create a new table
// @Description Create a new table for a restaurant
// @Tags Table
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param restaurant_id path string true "Restaurant ID"
// @Param table body Table true "Table data"
// @Success 201 {object} Table
// @Failure 400 {string} string "Invalid input"
// @Failure 404 {string} string "Restaurant not found"
// @Failure 500 {string} string "Error creating table"
// @Router /api/restaurant/{restaurant_id}/table [post]
func CreateTable(c *fiber.Ctx) error {
	username := c.Locals("username").(string)
	restaurantID := c.Params("restaurant_id")

	restaurant, err := verifyRestaurantOwnership(username, parseUint(restaurantID))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Restaurant not found",
		})
	}

	var request struct {
		TableNumber int `json:"table_number"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Invalid input",
		})
	}

	table := models.Table{
		RestaurantID: restaurant.ID,
		TableNumber:  request.TableNumber,
	}

	if err := database.DB.Create(&table).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Error creating table",
		})
	}

	// After creating the table, generate the QR code image
	frontendURL := fmt.Sprintf("http://localhost:5173/restaurant/%d/table/%d", restaurant.ID, table.ID)

	qrCode, err := utils.GenerateQRCode(frontendURL)
	if err != nil {
		// Log error but don't fail the operation
		fmt.Println("Error generating QR code:", err)
		// Set a fallback QR code URL if generation fails
		table.QRCodeURL = utils.GenerateFallbackQRCode(frontendURL)
	} else {
		table.QRCodeURL = qrCode
	}

	if err := database.DB.Save(&table).Error; err != nil {
		// Log error but don't fail the operation
		fmt.Println("Error updating table with QR code URL:", err)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    table,
		"error":   nil,
	})
}

// GetTables godoc
// @Summary Get all tables
// @Description Get all tables for a restaurant
// @Tags Table
// @Produce json
// @Security BearerAuth
// @Param restaurant_id path string true "Restaurant ID"
// @Success 200 {array} Table
// @Failure 404 {string} string "Restaurant not found"
// @Failure 500 {string} string "Error retrieving tables"
// @Router /api/restaurant/{restaurant_id}/table [get]
func GetTables(c *fiber.Ctx) error {
	username := c.Locals("username").(string)
	restaurantID := c.Params("restaurant_id")

	restaurant, err := verifyRestaurantOwnership(username, parseUint(restaurantID))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Restaurant not found",
		})
	}

	var tables []models.Table
	if err := database.DB.Where("restaurant_id = ?", restaurant.ID).Find(&tables).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Error retrieving tables",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    tables,
		"error":   nil,
	})
}

// UpdateTable godoc
// @Summary Update a table
// @Description Update a table
// @Tags Table
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param restaurant_id path string true "Restaurant ID"
// @Param id path string true "Table ID"
// @Param table body Table true "Table data"
// @Success 200 {object} Table
// @Failure 400 {string} string "Invalid input"
// @Failure 404 {string} string "Restaurant or table not found"
// @Failure 500 {string} string "Error updating table"
// @Router /api/restaurant/{restaurant_id}/table/{id} [put]
func UpdateTable(c *fiber.Ctx) error {
	username := c.Locals("username").(string)
	restaurantID := c.Params("restaurant_id")
	tableID := c.Params("id")

	restaurant, err := verifyRestaurantOwnership(username, parseUint(restaurantID))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Restaurant not found",
		})
	}

	var table models.Table
	if err := database.DB.Where("id = ? AND restaurant_id = ?", tableID, restaurant.ID).First(&table).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Table not found",
		})
	}

	var request struct {
		TableNumber int `json:"table_number"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Invalid input",
		})
	}

	table.TableNumber = request.TableNumber

	// Don't allow updating QRCodeURL from the frontend, regenerate it if necessary
	frontendURL := fmt.Sprintf("http://localhost:5173/restaurant/%d/table/%d", restaurant.ID, table.ID)

	qrCode, err := utils.GenerateQRCode(frontendURL)
	if err != nil {
		// Log error but don't fail the operation
		fmt.Println("Error generating QR code:", err)
		// Set a fallback QR code URL if generation fails
		table.QRCodeURL = utils.GenerateFallbackQRCode(frontendURL)
	} else {
		table.QRCodeURL = qrCode
	}

	if err := database.DB.Save(&table).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Error updating table",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    table,
		"error":   nil,
	})
}

// DeleteTable godoc
// @Summary Delete a table
// @Description Delete a table
// @Tags Table
// @Produce json
// @Security BearerAuth
// @Param restaurant_id path string true "Restaurant ID"
// @Param id path string true "Table ID"
// @Success 200 {object} string
// @Failure 404 {string} string "Restaurant or table not found"
// @Failure 500 {string} string "Error deleting table"
// @Router /api/restaurant/{restaurant_id}/table/{id} [delete]
func DeleteTable(c *fiber.Ctx) error {
	username := c.Locals("username").(string)
	restaurantID := c.Params("restaurant_id")
	tableID := c.Params("id")

	restaurant, err := verifyRestaurantOwnership(username, parseUint(restaurantID))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Restaurant not found",
		})
	}

	var table models.Table
	if err := database.DB.Where("id = ? AND restaurant_id = ?", tableID, restaurant.ID).First(&table).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Table not found",
		})
	}

	if err := database.DB.Delete(&table).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Error deleting table",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    "Table deleted successfully",
		"error":   nil,
	})
}

// GetAllUserTables godoc
// @Summary Get all user tables
// @Description Get all tables for all restaurants belonging to the user
// @Tags Table
// @Produce json
// @Security BearerAuth
// @Success 200 {array} Table
// @Failure 404 {string} string "User not found"
// @Failure 500 {string} string "Error retrieving tables"
// @Router /api/table [get]
func GetAllUserTables(c *fiber.Ctx) error {
	username := c.Locals("username").(string)

	var user models.User
	if err := database.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "User not found",
		})
	}

	// Get all restaurants for the user
	var restaurants []models.Restaurant
	if err := database.DB.Where("user_id = ?", user.ID).Find(&restaurants).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Error retrieving restaurants",
		})
	}

	// Extract restaurant IDs
	var restaurantIDs []uint
	for _, restaurant := range restaurants {
		restaurantIDs = append(restaurantIDs, restaurant.ID)
	}

	// If user has no restaurants, return empty array
	if len(restaurantIDs) == 0 {
		return c.JSON(fiber.Map{
			"success": true,
			"data":    []models.Table{},
			"error":   nil,
		})
	}

	// Get all tables for these restaurants
	var tables []models.Table
	if err := database.DB.Where("restaurant_id IN ?", restaurantIDs).Find(&tables).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Error retrieving tables",
		})
	}

	// Enhance table data with restaurant information
	var tablesWithRestaurantInfo []map[string]interface{}
	for _, table := range tables {
		// Regenerate QR code if the stored URL is empty or not base64 encoded
		qrCodeURL := table.QRCodeURL
		if qrCodeURL == "" {
			frontendURL := fmt.Sprintf("http://localhost:5173/restaurant/%d/table/%d", table.RestaurantID, table.ID)
			qrCode, err := utils.GenerateQRCode(frontendURL)
			if err != nil {
				// Log error and use fallback
				fmt.Println("Error generating QR code:", err)
				qrCodeURL = utils.GenerateFallbackQRCode(frontendURL)
			} else {
				qrCodeURL = qrCode
			}
		}

		tableMap := map[string]interface{}{
			"ID":           table.ID,
			"RestaurantID": table.RestaurantID,
			"TableNumber":  table.TableNumber,
			"QRCodeURL":    qrCodeURL,
		}

		// Find the restaurant for this table to add its name
		for _, restaurant := range restaurants {
			if restaurant.ID == table.RestaurantID {
				tableMap["RestaurantName"] = restaurant.Name
				break
			}
		}

		tablesWithRestaurantInfo = append(tablesWithRestaurantInfo, tableMap)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    tablesWithRestaurantInfo,
		"error":   nil,
	})
}

func parseUint(s string) uint {
	var id uint
	for _, c := range s {
		if c >= '0' && c <= '9' {
			id = id*10 + uint(c-'0')
		}
	}
	return id
}
