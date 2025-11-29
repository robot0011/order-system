package handler

import (
	"order-system/database"
	"order-system/models"
	"order-system/utils"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// CreateOrder godoc
// @Summary Create a new order
// @Description Create a new order for a restaurant
// @Tags Order
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param restaurant_id path string true "Restaurant ID"
// @Param order body Order true "Order data"
// @Success 201 {object} Order
// @Failure 400 {string} string "Invalid input"
// @Failure 404 {string} string "Restaurant, table, or menu item not found"
// @Failure 500 {string} string "Error creating order"
// @Router /api/restaurant/{restaurant_id}/order [post]
func CreateOrder(c *fiber.Ctx) error {
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
		TableID      uint   `json:"table_id"`
		CustomerName string `json:"customer_name"`
		OrderItems   []struct {
			MenuItemID          uint   `json:"menu_item_id"`
			Quantity            int    `json:"quantity"`
			SpecialInstructions string `json:"special_instructions"`
		} `json:"order_items"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Invalid input",
		})
	}

	// Verify table belongs to restaurant
	var table models.Table
	if err := database.DB.Where("id = ? AND restaurant_id = ?", request.TableID, restaurant.ID).First(&table).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Table not found",
		})
	}

	// Calculate total amount
	var totalAmount float64
	var orderItems []models.OrderItem

	for _, item := range request.OrderItems {
		var menuItem models.MenuItem
		if err := database.DB.Where("id = ? AND restaurant_id = ?", item.MenuItemID, restaurant.ID).First(&menuItem).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"data":    nil,
				"error":   "Menu item not found",
			})
		}

		quantity := item.Quantity
		if quantity <= 0 {
			quantity = 1
		}

		totalAmount += menuItem.Price * float64(quantity)

		orderItems = append(orderItems, models.OrderItem{
			MenuItemID:          item.MenuItemID,
			Quantity:            quantity,
			SpecialInstructions: item.SpecialInstructions,
		})
	}

	order := models.Order{
		TableID:      request.TableID,
		CustomerName: request.CustomerName,
		Status:       "pending",
		TotalAmount:  totalAmount,
		OrderItems:   orderItems,
	}

	if err := database.DB.Create(&order).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Error creating order",
		})
	}

	// Load order with items
	database.DB.Preload("OrderItems").First(&order, order.ID)

	orderResponse := buildOrderResponse(order, restaurant)
	globalOrderHub.publish("order_created", orderResponse)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    order,
		"error":   nil,
	})
}

// GetOrders godoc
// @Summary Get all orders
// @Description Get all orders for a restaurant
// @Tags Order
// @Produce json
// @Security BearerAuth
// @Param restaurant_id path string true "Restaurant ID"
// @Success 200 {array} Order
// @Failure 404 {string} string "Restaurant not found"
// @Failure 500 {string} string "Error retrieving orders"
// @Router /api/restaurant/{restaurant_id}/order [get]
func GetOrders(c *fiber.Ctx) error {
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

	// Get all table IDs for the restaurant
	var tables []models.Table
	database.DB.Where("restaurant_id = ?", restaurant.ID).Find(&tables)

	var tableIDs []uint
	for _, table := range tables {
		tableIDs = append(tableIDs, table.ID)
	}

	var orders []models.Order
	if len(tableIDs) > 0 {
		if err := database.DB.Where("table_id IN ?", tableIDs).Preload("OrderItems").Find(&orders).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"data":    nil,
				"error":   "Error retrieving orders",
			})
		}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    orders,
		"error":   nil,
	})
}

// GetOrder godoc
// @Summary Get order by ID
// @Description Get a single order by ID
// @Tags Order
// @Produce json
// @Security BearerAuth
// @Param restaurant_id path string true "Restaurant ID"
// @Param id path string true "Order ID"
// @Success 200 {object} Order
// @Failure 404 {string} string "Restaurant or order not found"
// @Failure 500 {string} string "Error retrieving order"
// @Router /api/restaurant/{restaurant_id}/order/{id} [get]
func GetOrder(c *fiber.Ctx) error {
	username := c.Locals("username").(string)
	restaurantID := c.Params("restaurant_id")
	orderID := c.Params("id")

	restaurant, err := verifyRestaurantOwnership(username, parseUint(restaurantID))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Restaurant not found",
		})
	}

	// Get all table IDs for the restaurant
	var tables []models.Table
	database.DB.Where("restaurant_id = ?", restaurant.ID).Find(&tables)

	var tableIDs []uint
	for _, table := range tables {
		tableIDs = append(tableIDs, table.ID)
	}

	var order models.Order
	if err := database.DB.Where("id = ? AND table_id IN ?", orderID, tableIDs).Preload("OrderItems").First(&order).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Order not found",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    order,
		"error":   nil,
	})
}

// UpdateOrderStatus godoc
// @Summary Update order status
// @Description Update the status of an order
// @Tags Order
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param restaurant_id path string true "Restaurant ID"
// @Param id path string true "Order ID"
// @Param status body OrderStatusUpdate true "Order status"
// @Success 200 {object} Order
// @Failure 400 {string} string "Invalid input"
// @Failure 404 {string} string "Restaurant or order not found"
// @Failure 500 {string} string "Error updating order"
// @Router /api/restaurant/{restaurant_id}/order/{id} [patch]
func UpdateOrderStatus(c *fiber.Ctx) error {
	username := c.Locals("username").(string)
	restaurantID := c.Params("restaurant_id")
	orderID := c.Params("id")

	restaurant, err := verifyRestaurantOwnership(username, parseUint(restaurantID))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Restaurant not found",
		})
	}

	// Get all table IDs for the restaurant
	var tables []models.Table
	database.DB.Where("restaurant_id = ?", restaurant.ID).Find(&tables)

	var tableIDs []uint
	for _, table := range tables {
		tableIDs = append(tableIDs, table.ID)
	}

	var order models.Order
	if err := database.DB.Where("id = ? AND table_id IN ?", orderID, tableIDs).First(&order).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Order not found",
		})
	}

	var request struct {
		Status string `json:"status"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Invalid input",
		})
	}

	// Map the simplified frontend status to internal status value
	internalStatus := utils.MapFrontendStatusToInternal(request.Status)
	order.Status = internalStatus

	if err := database.DB.Save(&order).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Error updating order",
		})
	}

	database.DB.Preload("OrderItems").First(&order, order.ID)
	orderResponse := buildOrderResponse(order, restaurant)
	globalOrderHub.publish("order_updated", orderResponse)

	return c.JSON(fiber.Map{
		"success": true,
		"data":    order,
		"error":   nil,
	})
}

// DeleteOrder godoc
// @Summary Delete an order
// @Description Delete an order
// @Tags Order
// @Produce json
// @Security BearerAuth
// @Param restaurant_id path string true "Restaurant ID"
// @Param id path string true "Order ID"
// @Success 200 {object} string
// @Failure 404 {string} string "Restaurant or order not found"
// @Failure 500 {string} string "Error deleting order"
// @Router /api/restaurant/{restaurant_id}/order/{id} [delete]
func DeleteOrder(c *fiber.Ctx) error {
	username := c.Locals("username").(string)
	restaurantID := c.Params("restaurant_id")
	orderID := c.Params("id")

	restaurant, err := verifyRestaurantOwnership(username, parseUint(restaurantID))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Restaurant not found",
		})
	}

	// Get all table IDs for the restaurant
	var tables []models.Table
	database.DB.Where("restaurant_id = ?", restaurant.ID).Find(&tables)

	var tableIDs []uint
	for _, table := range tables {
		tableIDs = append(tableIDs, table.ID)
	}

	var order models.Order
	if err := database.DB.Where("id = ? AND table_id IN ?", orderID, tableIDs).First(&order).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Order not found",
		})
	}

	// Delete order items first
	database.DB.Where("order_id = ?", order.ID).Delete(&models.OrderItem{})

	if err := database.DB.Delete(&order).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Error deleting order",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    "Order deleted successfully",
		"error":   nil,
	})
}

// CreatePublicOrder godoc
// @Summary Create a public order
// @Description Create a new order for a restaurant without authentication
// @Tags Order
// @Accept json
// @Produce json
// @Param restaurant_id path string true "Restaurant ID"
// @Param order body Order true "Order data"
// @Success 201 {object} Order
// @Failure 400 {string} string "Invalid input"
// @Failure 404 {string} string "Restaurant, table, or menu item not found"
// @Failure 500 {string} string "Error creating order"
// @Router /api/restaurants/{restaurant_id}/order [post]
func CreatePublicOrder(c *fiber.Ctx) error {
	restaurantID := c.Params("restaurant_id")

	// Verify restaurant exists
	var restaurant models.Restaurant
	if err := database.DB.First(&restaurant, restaurantID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Restaurant not found",
		})
	}

	var request struct {
		TableID      uint   `json:"table_id"`
		CustomerName string `json:"customer_name"`
		OrderItems   []struct {
			MenuItemID          uint   `json:"menu_item_id"`
			Quantity            int    `json:"quantity"`
			SpecialInstructions string `json:"special_instructions"`
		} `json:"order_items"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Invalid input",
		})
	}

	// Verify table belongs to restaurant
	var table models.Table
	if err := database.DB.Where("id = ? AND restaurant_id = ?", request.TableID, restaurant.ID).First(&table).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Table not found",
		})
	}

	var createdOrder models.Order
	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		var totalAmount float64
		var orderItems []models.OrderItem

		for _, item := range request.OrderItems {
			var menuItem models.MenuItem
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("id = ? AND restaurant_id = ?", item.MenuItemID, restaurant.ID).
				First(&menuItem).Error; err != nil {
				return fiber.NewError(fiber.StatusNotFound, "Menu item not found")
			}

			quantity := item.Quantity
			if quantity <= 0 {
				quantity = 1
			}

			if menuItem.Quantity < quantity {
				return fiber.NewError(fiber.StatusBadRequest, "Insufficient quantity for item: "+menuItem.Name)
			}

			totalAmount += menuItem.Price * float64(quantity)

			menuItem.Quantity -= quantity
			if menuItem.Quantity < 0 {
				menuItem.Quantity = 0
			}

			if err := tx.Save(&menuItem).Error; err != nil {
				return err
			}

			orderItems = append(orderItems, models.OrderItem{
				MenuItemID:          item.MenuItemID,
				Quantity:            quantity,
				SpecialInstructions: item.SpecialInstructions,
			})
		}

		createdOrder = models.Order{
			TableID:      request.TableID,
			CustomerName: request.CustomerName,
			Status:       "pending",
			TotalAmount:  totalAmount,
			OrderItems:   orderItems,
		}

		if err := tx.Create(&createdOrder).Error; err != nil {
			return err
		}

		return tx.Preload("OrderItems").First(&createdOrder, createdOrder.ID).Error
	}); err != nil {
		if fiberErr, ok := err.(*fiber.Error); ok {
			return c.Status(fiberErr.Code).JSON(fiber.Map{
				"success": false,
				"data":    nil,
				"error":   fiberErr.Message,
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Error creating order",
		})
	}

	orderResponse := buildOrderResponse(createdOrder, &restaurant)
	globalOrderHub.publish("order_created", orderResponse)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    createdOrder,
		"error":   nil,
	})
}

func buildOrderResponse(order models.Order, restaurant *models.Restaurant) OrderResponse {
	updatedOrder := order
	updatedOrder.Status = utils.MapInternalStatusToFrontend(order.Status)

	// Convert models.Order to handler.Order
	handlerOrder := Order{
		ID:           updatedOrder.ID,
		TableID:      updatedOrder.TableID,
		CustomerName: updatedOrder.CustomerName,
		Status:       updatedOrder.Status,
		TotalAmount:  updatedOrder.TotalAmount,
		OrderItems:   make([]OrderItem, len(updatedOrder.OrderItems)),
	}

	// Convert order items
	for i, item := range updatedOrder.OrderItems {
		handlerOrder.OrderItems[i] = OrderItem{
			ID:                  item.ID,
			OrderID:             item.OrderID,
			MenuItemID:          item.MenuItemID,
			Quantity:            item.Quantity,
			SpecialInstructions: item.SpecialInstructions,
		}
	}

	return OrderResponse{
		Order:          handlerOrder,
		RestaurantName: restaurant.Name,
		RestaurantID:   restaurant.ID,
	}
}

// GetAllUserOrders godoc
// @Summary Get all user orders
// @Description Get all orders for all restaurants belonging to the user
// @Tags Order
// @Produce json
// @Security BearerAuth
// @Success 200 {array} OrderResponse
// @Failure 404 {string} string "User not found"
// @Failure 500 {string} string "Error retrieving orders"
// @Router /api/order [get]
func GetAllUserOrders(c *fiber.Ctx) error {
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
			"data":    []OrderResponse{},
			"error":   nil,
		})
	}

	// Get all tables for these restaurants to get the table IDs
	var tables []models.Table
	if err := database.DB.Where("restaurant_id IN ?", restaurantIDs).Find(&tables).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Error retrieving tables",
		})
	}

	// Extract table IDs
	var tableIDs []uint
	for _, table := range tables {
		tableIDs = append(tableIDs, table.ID)
	}

	// If user has no tables, return empty array
	if len(tableIDs) == 0 {
		return c.JSON(fiber.Map{
			"success": true,
			"data":    []OrderResponse{},
			"error":   nil,
		})
	}

	// Get all orders for these tables
	var orders []models.Order
	if err := database.DB.Where("table_id IN ?", tableIDs).Preload("OrderItems").Preload("OrderItems.MenuItem").Find(&orders).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Error retrieving orders",
		})
	}

	// Convert orders to OrderResponse with restaurant name and ID
	var orderResponses []OrderResponse
	for _, order := range orders {
		restaurantName := ""
		var restaurantID uint = 0
		for _, restaurant := range restaurants {
			if isRestaurantForTable(restaurant.ID, order.TableID, tables) {
				restaurantName = restaurant.Name
				restaurantID = restaurant.ID
				break
			}
		}

		// Map internal status to simplified frontend status
		simplifiedStatus := utils.MapInternalStatusToFrontend(order.Status)

		// Create a copy of the order with the simplified status
		updatedOrder := order
		updatedOrder.Status = simplifiedStatus

		// Convert models.Order to handler.Order
		handlerOrder := Order{
			ID:           updatedOrder.ID,
			TableID:      updatedOrder.TableID,
			CustomerName: updatedOrder.CustomerName,
			Status:       updatedOrder.Status,
			TotalAmount:  updatedOrder.TotalAmount,
			OrderItems:   make([]OrderItem, len(updatedOrder.OrderItems)),
		}

		// Convert order items
		for i, item := range updatedOrder.OrderItems {
			handlerOrder.OrderItems[i] = OrderItem{
				ID:                  item.ID,
				OrderID:             item.OrderID,
				MenuItemID:          item.MenuItemID,
				Quantity:            item.Quantity,
				SpecialInstructions: item.SpecialInstructions,
			}
		}

		orderResponses = append(orderResponses, OrderResponse{
			Order:          handlerOrder,
			RestaurantName: restaurantName,
			RestaurantID:   restaurantID,
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    orderResponses,
		"error":   nil,
	})
}

// Helper function to determine if a restaurant owns a specific table
func isRestaurantForTable(restaurantID uint, tableID uint, tables []models.Table) bool {
	for _, table := range tables {
		if table.ID == tableID {
			return table.RestaurantID == restaurantID
		}
	}
	return false
}
