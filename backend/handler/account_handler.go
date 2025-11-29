package handler

import (
	"errors"
	"fmt"
	"order-system/database"
	"order-system/models"
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

	var dbUser models.User
	err := database.DB.Where("username = ?", loginRequest.Username).Preload("Restaurants").First(&dbUser).Error
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Invalid username or password",
		})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(loginRequest.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Invalid username or password",
		})
	}

	// Generate Access and Refresh Tokens
	accessToken, err := generateAccessToken(loginRequest.Username)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Error generating access token",
		})
	}

	refreshToken, err := generateRefreshToken(loginRequest.Username)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Error generating refresh token",
		})
	}

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

	// Return response
	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"user_id":       dbUser.ID,
			"username":      dbUser.Username,
			"email":         dbUser.Email,
			"role":          dbUser.Role,
			"access_token":  accessToken,
			"refresh_token": refreshToken,
			"restaurant":    restaurantData,
		},
		"error": nil,
	})
}

// Middleware to protect routes using Access Token
func ProtectRoute(c *fiber.Ctx) error {
	tokenString, err := extractBearerToken(c.Get("Authorization"))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   err.Error(),
		})
	}

	username, err := ValidateAccessToken(tokenString)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   err.Error(),
		})
	}

	c.Locals("username", username)

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
	var refreshRequest struct {
		RefreshToken string `json:"refresh_token"`
	}

	// Parse the refresh token from the body
	if err := c.BodyParser(&refreshRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Invalid input",
		})
	}

	// Validate the refresh token
	token, err := jwt.Parse(refreshRequest.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(refreshSecretKey), nil
	})

	// Check if the token is valid and not expired
	if err != nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Invalid or expired refresh token",
		})
	}

	// Extract username from the refresh token's claims
	claims := token.Claims.(jwt.MapClaims)
	username := claims["username"].(string)

	// Generate a new access token
	accessToken, err := generateAccessToken(username)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Error generating new access token",
		})
	}

	// Optionally, generate a new refresh token (if you want the refresh token to change)
	refreshToken, err := generateRefreshToken(username)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"data":    nil,
			"error":   "Error generating new refresh token",
		})
	}

	// Return the new tokens
	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
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
