package handler

import (
	"order-system/database"
	"order-system/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var secretKey = "mysecretkey"               // In production, use an environment variable for the secret key.
var refreshSecretKey = "myrefreshsecretkey" // For refresh tokens

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

// Handler for registration (creating new user)
func Register(c *fiber.Ctx) error {
	var registerRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Name     string `json:"name"`
		Email    string `json:"email"`
	}

	// Parse the registration data
	if err := c.BodyParser(&registerRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid input")
	}

	// Check if user already exists
	var existingUser models.User
	err := database.DB.Where("username = ?", registerRequest.Username).First(&existingUser).Error
	if err == nil {
		return c.Status(fiber.StatusConflict).SendString("Username already taken")
	} else if err != nil && err.Error() != "record not found" && err != gorm.ErrRecordNotFound {
		return c.Status(fiber.StatusInternalServerError).SendString("Database error")
	}

	// Check if email already exists
	var existingEmailUser models.User
	err = database.DB.Where("email = ?", registerRequest.Email).First(&existingEmailUser).Error
	if err == nil {
		return c.Status(fiber.StatusConflict).SendString("Email already registered")
	} else if err != nil && err.Error() != "record not found" && err != gorm.ErrRecordNotFound {
		return c.Status(fiber.StatusInternalServerError).SendString("Database error")
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(registerRequest.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error hashing password")
	}

	// Create a new user and save to the database
	user := models.User{
		Username: registerRequest.Username,
		Password: string(hashedPassword),
		Name:     registerRequest.Name,
		Email:    registerRequest.Email,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error creating user")
	}

	// Return success response
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User created successfully",
		"user":    user,
	})
}

// Handler for login (creating JWT and refresh token)
func Login(c *fiber.Ctx) error {
	var loginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&loginRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid input")
	}

	var dbUser models.User
	err := database.DB.Where("username = ?", loginRequest.Username).First(&dbUser).Error
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).SendString("Invalid username or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(loginRequest.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).SendString("Invalid username or password")
	}

	// Generate Access and Refresh Tokens
	accessToken, err := generateAccessToken(loginRequest.Username)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error generating access token")
	}

	refreshToken, err := generateRefreshToken(loginRequest.Username)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error generating refresh token")
	}

	// Return both tokens to the user
	return c.JSON(fiber.Map{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

// Middleware to protect routes using Access Token
func ProtectRoute(c *fiber.Ctx) error {
	tokenString := c.Get("Authorization")

	if len(tokenString) < 7 || tokenString[:7] != "Bearer " {
		return c.Status(fiber.StatusUnauthorized).SendString("Missing or malformed token")
	}
	tokenString = tokenString[7:] // Remove the "Bearer " part

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(secretKey), nil
	})

	if err != nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).SendString("Invalid or expired token")
	}

	claims := token.Claims.(jwt.MapClaims)
	username := claims["username"].(string)

	c.Locals("username", username)

	return c.Next()
}

// Protected route to return user profile
func Profile(c *fiber.Ctx) error {
	username := c.Locals("username").(string)

	var user models.User
	err := database.DB.Where("username = ?", username).First(&user).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Could not retrieve user profile")
	}

	return c.JSON(fiber.Map{
		"name":  user.Name,
		"email": user.Email,
	})
}

// Handler for refresh token (get a new access token using the refresh token)
func RefreshToken(c *fiber.Ctx) error {
	var refreshRequest struct {
		RefreshToken string `json:"refresh_token"`
	}

	// Parse the refresh token from the body
	if err := c.BodyParser(&refreshRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid input")
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
		return c.Status(fiber.StatusUnauthorized).SendString("Invalid or expired refresh token")
	}

	// Extract username from the refresh token's claims
	claims := token.Claims.(jwt.MapClaims)
	username := claims["username"].(string)

	// Generate a new access token
	accessToken, err := generateAccessToken(username)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error generating new access token")
	}

	// Optionally, generate a new refresh token (if you want the refresh token to change)
	refreshToken, err := generateRefreshToken(username)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error generating new refresh token")
	}

	// Return the new tokens
	return c.JSON(fiber.Map{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

func GetAllUsers(c *fiber.Ctx) error {
	var users []models.User
	err := database.DB.Find(&users).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Could not retrieve users")
	}

	return c.JSON(users)
}
