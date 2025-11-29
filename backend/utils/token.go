package utils

import (
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// GenerateSecureAccessToken generates a secure access token with additional claims
func GenerateSecureAccessToken(userID uint, username string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = userID
	claims["username"] = username
	claims["exp"] = time.Now().Add(time.Minute * 15).Unix() // Access token expires in 15 minutes
	claims["type"] = "access"

	t, err := token.SignedString([]byte(getSecretKey()))
	if err != nil {
		return "", err
	}

	return t, nil
}

// GenerateSecureRefreshToken generates a secure refresh token with additional claims
func GenerateSecureRefreshToken(userID uint, username string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = userID
	claims["username"] = username
	claims["exp"] = time.Now().Add(time.Hour * 24 * 30).Unix() // Refresh token expires in 30 days
	claims["type"] = "refresh"

	t, err := token.SignedString([]byte(getRefreshSecretKey()))
	if err != nil {
		return "", err
	}

	return t, nil
}

// ValidateAccessToken validates the access token
func ValidateAccessToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(getSecretKey()), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	// Check if token type is access or websocket
	tokenType := claims["type"]
	if tokenType != "access" && tokenType != "websocket" {
		return nil, jwt.ErrSignatureInvalid
	}

	return claims, nil
}

// ValidateRefreshToken validates the refresh token
func ValidateRefreshToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(getRefreshSecretKey()), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	// Check if token type is refresh
	if claims["type"] != "refresh" {
		return nil, jwt.ErrSignatureInvalid
	}

	return claims, nil
}

// SetSecureCookie sets a secure cookie with HttpOnly and Secure flags
func SetSecureCookie(c *fiber.Ctx, name, value string, maxAge int) {
	c.Cookie(&fiber.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   maxAge,
		HTTPOnly: true,
		Secure:   true, // Only send over HTTPS
		SameSite: "Strict", // Prevent CSRF
		Path:     "/",
	})
}

// ClearSecureCookie clears a secure cookie
func ClearSecureCookie(c *fiber.Ctx, name string) {
	c.Cookie(&fiber.Cookie{
		Name:     name,
		Value:    "",
		MaxAge:   -1,
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Strict",
		Path:     "/",
	})
}

// getSecretKey gets the JWT secret key from environment variable
func getSecretKey() string {
	// In production, use an environment variable for the secret key
	secret := getEnvOrDefault("JWT_SECRET", "very_long_secret_key_for_production")
	return secret
}

// getRefreshSecretKey gets the refresh token secret key from environment variable
func getRefreshSecretKey() string {
	refreshSecret := getEnvOrDefault("JWT_REFRESH_SECRET", "very_long_refresh_secret_key_for_production")
	return refreshSecret
}

// GenerateSecureWebSocketToken generates a short-lived token specifically for WebSocket connections
func GenerateSecureWebSocketToken(userID uint, username string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = userID
	claims["username"] = username
	claims["exp"] = time.Now().Add(time.Minute * 5).Unix() // WebSocket token expires in 5 minutes
	claims["type"] = "websocket"

	t, err := token.SignedString([]byte(getSecretKey()))
	if err != nil {
		return "", err
	}

	return t, nil
}

// getEnvOrDefault gets an environment variable or returns a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}