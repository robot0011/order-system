package utils

import (
	"github.com/gofiber/fiber/v2"
)

// APIResponse represents the standard response format
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

// SuccessResponse returns a successful API response
func SuccessResponse(data interface{}) APIResponse {
	return APIResponse{
		Success: true,
		Data:    data,
		Error:   nil,
	}
}

// ErrorResponse returns an error API response
func ErrorResponse(error interface{}) APIResponse {
	return APIResponse{
		Success: false,
		Data:    nil,
		Error:   error,
	}
}

// SendSuccess sends a successful JSON response with the provided data
func SendSuccess(c *fiber.Ctx, data interface{}) error {
	return c.JSON(SuccessResponse(data))
}

// SendSuccessWithStatus sends a successful JSON response with data and status code
func SendSuccessWithStatus(c *fiber.Ctx, data interface{}, statusCode int) error {
	return c.Status(statusCode).JSON(SuccessResponse(data))
}

// SendError sends an error JSON response with the provided error message
func SendError(c *fiber.Ctx, statusCode int, error interface{}) error {
	return c.Status(statusCode).JSON(ErrorResponse(error))
}

// SendResponse sends a custom API response
func SendResponse(c *fiber.Ctx, success bool, data interface{}, error interface{}, statusCode int) error {
	response := APIResponse{
		Success: success,
		Data:    data,
		Error:   error,
	}
	return c.Status(statusCode).JSON(response)
}