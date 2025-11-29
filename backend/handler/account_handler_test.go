package handler

import (
	"net/http"
	"net/http/httptest"
	"order-system/utils"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateAccessToken(t *testing.T) {
	tokenString, err := generateAccessToken("tester")
	if err != nil {
		t.Fatalf("generateAccessToken returned error: %v", err)
	}
	if tokenString == "" {
		t.Fatal("generateAccessToken returned empty token")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}

	claims := token.Claims.(jwt.MapClaims)
	if claims["username"] != "tester" {
		t.Fatalf("expected username tester, got %v", claims["username"])
	}
	if exp, ok := claims["exp"].(float64); ok {
		if time.Unix(int64(exp), 0).Before(time.Now()) {
			t.Fatal("token already expired")
		}
	} else {
		t.Fatal("missing exp claim")
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	tokenString, err := generateRefreshToken("tester")
	if err != nil {
		t.Fatalf("generateRefreshToken returned error: %v", err)
	}
	if tokenString == "" {
		t.Fatal("generateRefreshToken returned empty token")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(refreshSecretKey), nil
	})
	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}

	claims := token.Claims.(jwt.MapClaims)
	if claims["username"] != "tester" {
		t.Fatalf("expected username tester, got %v", claims["username"])
	}
	if exp, ok := claims["exp"].(float64); ok {
		if time.Unix(int64(exp), 0).Before(time.Now()) {
			t.Fatal("token already expired")
		}
	} else {
		t.Fatal("missing exp claim")
	}
}

func TestProtectRouteSuccess(t *testing.T) {
	app := fiber.New()
	app.Use(ProtectRoute)
	app.Get("/protected", func(c *fiber.Ctx) error {
		if c.Locals("username") != "tester" {
			return fiber.ErrUnauthorized
		}
		return c.SendStatus(fiber.StatusOK)
	})

	// Use the new secure token generation
	token, err := utils.GenerateSecureAccessToken(1, "tester")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	// Set the token as a cookie instead of header
	req.AddCookie(&http.Cookie{
		Name:  "access_token",
		Value: token,
	})

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("fiber app test failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestProtectRouteMissingToken(t *testing.T) {
	app := fiber.New()
	app.Use(ProtectRoute)
	app.Get("/protected", func(c *fiber.Ctx) error {
		return fiber.ErrUnauthorized
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("fiber app test failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", resp.StatusCode)
	}
}
