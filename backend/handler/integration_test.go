package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"order-system/database"
	"order-system/models"
	"os"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func init() {
	// Change to project root so .env can be found
	os.Chdir("..")
}

func setupTestApp() *fiber.App {
	app := fiber.New()

	// User routes
	api := app.Group("/api")
	user := api.Group("/user")
	user.Post("/register", Register)
	user.Post("/login", Login)
	user.Get("/profile", ProtectRoute, Profile)
	user.Delete("/", ProtectRoute, DeleteUser)

	// Restaurant routes
	restaurant := api.Group("/restaurant", ProtectRoute)
	restaurant.Post("/", CreateRestaurant)
	restaurant.Get("/", GetRestaurants)
	restaurant.Get("/:id", GetRestaurantByID)
	restaurant.Put("/:id", UpdateRestaurant)
	restaurant.Delete("/:id", DeleteRestaurant)

	return app
}

func TestFullFlow(t *testing.T) {
	// Initialize test database
	database.ConnectDB()

	app := setupTestApp()

	// Cleanup any existing test data first (including soft-deleted)
	var existingUser models.User
	if err := database.DB.Unscoped().Where("username = ?", "testuser_integration").First(&existingUser).Error; err == nil {
		database.DB.Unscoped().Where("user_id = ?", existingUser.ID).Delete(&models.Restaurant{})
		database.DB.Unscoped().Delete(&existingUser)
	}

	// Dummy data
	testUser := map[string]string{
		"username": "testuser_integration",
		"password": "testpassword123",
		"email":    "testintegration@example.com",
		"role":     "owner",
	}

	testRestaurant := map[string]string{
		"name":         "Test Restaurant",
		"address":      "123 Test Street",
		"phone_number": "123-456-7890",
		"logo_url":     "https://example.com/logo.png",
	}

	var accessToken string
	var restaurantID uint

	// 1. Test Register
	t.Run("Register", func(t *testing.T) {
		body, _ := json.Marshal(testUser)
		req := httptest.NewRequest("POST", "/api/user/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 201, resp.StatusCode)
	})

	// 2. Test Login
	t.Run("Login", func(t *testing.T) {
		loginData := map[string]string{
			"username": testUser["username"],
			"password": testUser["password"],
		}
		body, _ := json.Marshal(loginData)
		req := httptest.NewRequest("POST", "/api/user/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result struct {
			Success bool `json:"success"`
			Data    struct {
				AccessToken string `json:"access_token"`
				// RefreshToken string `json:"refresh_token"`
			} `json:"data"`
			Error interface{} `json:"error"`
		}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.True(t, result.Success, "Expected success to be true")
		assert.Nil(t, result.Error, "Expected error to be nil")
		accessToken = result.Data.AccessToken
		assert.NotEmpty(t, accessToken)
	})

	// 3. Test Get Profile
	t.Run("GetProfile", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/user/profile", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})

	// 4. Test Create Restaurant
	t.Run("CreateRestaurant", func(t *testing.T) {
		body, _ := json.Marshal(testRestaurant)
		req := httptest.NewRequest("POST", "/api/restaurant/", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 201, resp.StatusCode)

		var response struct {
			Success bool            `json:"success"`
			Data    models.Restaurant `json:"data"`
			Error   interface{}       `json:"error"`
		}
		json.NewDecoder(resp.Body).Decode(&response)
		assert.True(t, response.Success, "Expected success to be true")
		assert.Nil(t, response.Error, "Expected error to be nil")
		restaurantID = response.Data.ID
		assert.NotZero(t, restaurantID)
	})

	// 5. Test Get Restaurant
	t.Run("GetRestaurant", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/restaurant/", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})

	// 6. Test Update Restaurant
	t.Run("UpdateRestaurant", func(t *testing.T) {
		updatedRestaurant := map[string]string{
			"name":         "Updated Restaurant",
			"address":      "456 Updated Street",
			"phone_number": "999-999-9999",
			"logo_url":     "https://example.com/newlogo.png",
		}
		body, _ := json.Marshal(updatedRestaurant)
		req := httptest.NewRequest("PUT", fmt.Sprintf("/api/restaurant/%d", restaurantID), bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})

	// 7. Test Delete Restaurant
	t.Run("DeleteRestaurant", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/restaurant/%d", restaurantID), nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})

	// 8. Cleanup - Delete User
	t.Run("DeleteUser", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/user/", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})

	// Verify cleanup - user should not exist
	var deletedUser models.User
	err := database.DB.Where("username = ?", testUser["username"]).First(&deletedUser).Error
	assert.Error(t, err) // Should error because user is deleted
}
