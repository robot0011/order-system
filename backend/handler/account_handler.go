package handler

import (
	"errors"
	"fmt"
	"order-system/database"
	"order-system/models"
	"order-system/utils"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// HealthCheck godoc
// @Summary Health check endpoint
// @Description Check if the API is running
// @Tags Health
// @Produce json
// @Success 200 {object} utils.APIResponse
// @Router /health [get]
func HealthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"status":  "ok",
			"message": "Welcome to the Order-System API",
		},
		"error": nil,
	})
}

var secretKey = getEnvOrDefault("JWT_SECRET", "mysecretkey")                       // In production, use an environment variable for the secret key.
var refreshSecretKey = getEnvOrDefault("JWT_REFRESH_SECRET", "myrefreshsecretkey") // For refresh tokens

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Generate Access Token
func generateAccessToken(username string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["username"] = username
	claims["exp"] = time.Now().Add(time.Minute * 15).Unix() // Access token expires in 15 minutes

	t, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return t, nil
}

// Generate Refresh Token
func generateRefreshToken(username string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["username"] = username
	claims["exp"] = time.Now().Add(time.Hour * 24 * 30).Unix() // Refresh token expires in 30 days

	t, err := token.SignedString([]byte(refreshSecretKey))
	if err != nil {
		return "", err
	}

	return t, nil
}

// Register godoc
// @Summary Register a new user
// @Description Register a new user account
// @Tags User
// @Accept json
// @Produce json
// @Param user body RegisterRequest true "User registration data"
// @Success 201 {object} fiber.Map
// @Failure 400 {string} string "Invalid input"
// @Failure 409 {string} string "Username or email already taken"
// @Router /api/user/register [post]
func Register(c *fiber.Ctx) error {
	var registerRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
		Role     string `json:"role"`
	}

	// Parse the registration data
	if err := c.BodyParser(&registerRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Invalid input",
		})
	}

	// Check if user already exists
	var existingUser models.User
	err := database.DB.Where("username = ?", registerRequest.Username).First(&existingUser).Error
	if err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Username already taken",
		})
	} else if err != gorm.ErrRecordNotFound {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Database error",
		})
	}

	// Check if email already exists
	var existingEmailUser models.User
	err = database.DB.Where("email = ?", registerRequest.Email).First(&existingEmailUser).Error
	if err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Email already registered",
		})
	} else if err != gorm.ErrRecordNotFound {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Database error",
		})
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(registerRequest.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Error hashing password",
		})
	}

	// Create a new user and save to the database
	role := registerRequest.Role
	if role == "" {
		role = "owner"
	}
	user := models.User{
		Username: registerRequest.Username,
		Password: string(hashedPassword),
		Email:    registerRequest.Email,
		Role:     role,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Error creating user",
		})
	}

	// Return success response
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"user_id":  user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
		},
		"error": nil,
	})
}

// Login godoc
// @Summary User login
// @Description Login with username and password to get JWT tokens
// @Tags User
// @Accept json
// @Produce json
// @Param credentials body LoginRequest true "Login credentials"
// @Success 200 {object} fiber.Map
// @Failure 401 {string} string "Invalid username or password"
// @Failure 429 {string} string "Rate limit exceeded"
// @Router /api/user/login [post]
func Login(c *fiber.Ctx) error {
	var loginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&loginRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Invalid input",
		})
	}

	// Check brute force protection
	if !utils.CheckBruteForce(loginRequest.Username) {
		return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Too many failed login attempts. Account locked temporarily.",
		})
	}

	var dbUser models.User
	err := database.DB.Where("username = ?", loginRequest.Username).Preload("Restaurants").First(&dbUser).Error
	if err != nil {
		// Return generic error to prevent username enumeration
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Invalid username or password",
		})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(loginRequest.Password)); err != nil {
		// Login failed, don't record successful login
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Invalid username or password",
		})
	}

	// Login successful, record this to reset brute force attempts
	utils.RecordSuccessfulLogin(loginRequest.Username)

	// Generate Secure Access and Refresh Tokens
	accessToken, err := utils.GenerateSecureAccessToken(dbUser.ID, loginRequest.Username)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Error generating access token",
		})
	}

	refreshToken, err := utils.GenerateSecureRefreshToken(dbUser.ID, loginRequest.Username)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Error generating refresh token",
		})
	}

	// Set tokens as HttpOnly, Secure cookies
	utils.SetSecureCookie(c, "access_token", accessToken, 15*60) // 15 minutes
	utils.SetSecureCookie(c, "refresh_token", refreshToken, 30*24*60*60) // 30 days

	// Build restaurant response (nil if no restaurant)
	var restaurantData interface{}
	if len(dbUser.Restaurants) > 0 {
		firstRestaurant := dbUser.Restaurants[0]
		restaurantData = fiber.Map{
			"restaurant_id": firstRestaurant.ID,
			"name":          firstRestaurant.Name,
			"address":       firstRestaurant.Address,
			"phone_number":  firstRestaurant.PhoneNumber,
		}
	}

	// Return response without tokens in the body (they're in cookies)
	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"user_id":    dbUser.ID,
			"username":   dbUser.Username,
			"email":      dbUser.Email,
			"role":       dbUser.Role,
			"restaurant": restaurantData,
		},
		"error": nil,
	})
}

// Middleware to protect routes using Access Token from cookie
func ProtectRoute(c *fiber.Ctx) error {
	// Try to get token from cookie first, then from header as fallback
	tokenString := c.Cookies("access_token")
	if tokenString == "" {
		// Fallback to Authorization header for development/testing
		authHeader := c.Get("Authorization")
		if authHeader != "" {
			tokenString, _ = extractBearerToken(authHeader)
		}
	}

	if tokenString == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "No access token provided",
		})
	}

	claims, err := utils.ValidateAccessToken(tokenString)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Invalid or expired access token",
		})
	}

	// Store user info in context
	username, ok := claims["username"].(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Invalid token claims",
		})
	}

	// Get user_id from claims
	userID, ok := claims["user_id"].(float64)  // JWT numbers are float64
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Invalid token user ID",
		})
	}

	c.Locals("username", username)
	c.Locals("user_id", userID)

	return c.Next()
}

// Profile godoc
// @Summary Get user profile
// @Description Get the profile of the authenticated user
// @Tags User
// @Produce json
// @Security BearerAuth
// @Success 200 {object} User
// @Failure 401 {string} string "Unauthorized"
// @Router /api/user/profile [get]
func Profile(c *fiber.Ctx) error {
	username := c.Locals("username").(string)

	var user models.User
	err := database.DB.Where("username = ?", username).First(&user).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Could not retrieve user profile",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
		},
		"error": nil,
	})
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Get new access token using refresh token
// @Tags User
// @Accept json
// @Produce json
// @Param refresh body RefreshRequest true "Refresh token"
// @Success 200 {object} TokenResponse
// @Failure 401 {string} string "Invalid or expired refresh token"
// @Router /api/user/refresh [post]
func RefreshToken(c *fiber.Ctx) error {
	// Get refresh token from cookie
	refreshTokenString := c.Cookies("refresh_token")
	if refreshTokenString == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "No refresh token provided",
		})
	}

	// Validate the refresh token
	claims, err := utils.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		utils.ClearSecureCookie(c, "refresh_token") // Clear invalid token
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Invalid or expired refresh token",
		})
	}

	// Extract user info from the refresh token's claims
	username, ok := claims["username"].(string)
	userID, ok2 := claims["user_id"].(float64)
	if !ok || !ok2 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Invalid token claims",
		})
	}

	// Generate new access and refresh tokens
	newAccessToken, err := utils.GenerateSecureAccessToken(uint(userID), username)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Error generating new access token",
		})
	}

	newRefreshToken, err := utils.GenerateSecureRefreshToken(uint(userID), username)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Error generating new refresh token",
		})
	}

	// Set the new tokens as HttpOnly, Secure cookies
	utils.SetSecureCookie(c, "access_token", newAccessToken, 15*60) // 15 minutes
	utils.SetSecureCookie(c, "refresh_token", newRefreshToken, 30*24*60*60) // 30 days

	// Return success response
	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"message": "Tokens refreshed successfully",
		},
		"error": nil,
	})
}

func extractBearerToken(raw string) (string, error) {
	if raw == "" {
		return "", errors.New("missing token")
	}
	if len(raw) >= 7 && strings.ToLower(raw[:7]) == "bearer " {
		return raw[7:], nil
	}
	return raw, nil
}

func ValidateAccessToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(secretKey), nil
	})
	if err != nil {
		return "", errors.New("invalid or expired token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", errors.New("invalid or expired token")
	}
	username, ok := claims["username"].(string)
	if !ok {
		return "", errors.New("invalid token claims")
	}
	return username, nil
}

// GetAllUsers godoc
// @Summary Get all users
// @Description Get all registered users
// @Tags User
// @Produce json
// @Success 200 {array} User
// @Failure 500 {string} string "Could not retrieve users"
// @Router /api/user/ [get]
func GetAllUsers(c *fiber.Ctx) error {
	var users []models.User
	err := database.DB.Find(&users).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Could not retrieve users",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    users,
		"error":   nil,
	})
}

// DeleteUser godoc
// @Summary Delete user
// @Description Delete the authenticated user
// @Tags User
// @Produce json
// @Security BearerAuth
// @Success 200 {object} string
// @Failure 500 {string} string "Could not retrieve or delete user"
// @Router /api/user/ [delete]
func DeleteUser(c *fiber.Ctx) error {
	username := c.Locals("username").(string)

	var user models.User
	err := database.DB.Where("username = ?", username).First(&user).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Could not retrieve user",
		})
	}
	if err := database.DB.Delete(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Could not delete user",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    "User deleted successfully",
		"error":   nil,
	})
}

// Logout godoc
// @Summary User logout
// @Description Clear user's tokens
// @Tags User
// @Produce json
// @Success 200 {object} fiber.Map
// @Failure 500 {object} fiber.Map
// @Router /api/user/logout [post]
func Logout(c *fiber.Ctx) error {
	// Clear the secure tokens from cookies
	utils.ClearSecureCookie(c, "access_token")
	utils.ClearSecureCookie(c, "refresh_token")

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"message": "Successfully logged out",
		},
		"error": nil,
	})
}

// GetWebSocketToken godoc
// @Summary Get temporary WebSocket token
// @Description Get a temporary token for WebSocket authentication
// @Tags User
// @Produce json
// @Success 200 {object} fiber.Map
// @Failure 401 {object} fiber.Map
// @Router /api/user/websocket-token [get]
func GetWebSocketToken(c *fiber.Ctx) error {
	// Get username from context (set by ProtectRoute middleware)
	username, ok := c.Locals("username").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Invalid user context",
		})
	}

	// The user_id from JWT claims comes as float64, need to convert properly
	userIDFloat, ok := c.Locals("user_id").(float64)
	if !ok {
		// Try to get it as uint if it was set directly
		if userIDVal, ok := c.Locals("user_id").(uint); ok {
			return generateWebSocketTokenResponse(c, userIDVal, username)
		}
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Invalid user ID context",
		})
	}

	return generateWebSocketTokenResponse(c, uint(userIDFloat), username)
}

// Helper function to generate the WebSocket token response
func generateWebSocketTokenResponse(c *fiber.Ctx, userID uint, username string) error {
	// Generate a short-lived token specifically for WebSocket use
	// This token will have a short expiration and specific purpose
	websocketToken, err := utils.GenerateSecureWebSocketToken(userID, username)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Error generating WebSocket token",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"websocket_token": websocketToken,
		},
		"error": nil,
	})
}
