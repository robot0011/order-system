package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestHealthCheck(t *testing.T) {
	app := fiber.New()
	app.Get("/health", healthCheck)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("fiber app test failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Check the success field
	success, ok := body["success"].(bool)
	if !ok || !success {
		t.Fatalf("expected success to be true, got %v", body["success"])
	}

	// Check the data field
	data, ok := body["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data to be a map, got %v", body["data"])
	}

	if data["status"] != "ok" {
		t.Fatalf("expected status ok, got %s", data["status"])
	}
	if data["message"] == "" {
		t.Fatalf("expected message, got empty string")
	}

	// Check the error field
	if body["error"] != nil {
		t.Fatalf("expected error to be nil, got %v", body["error"])
	}
}
