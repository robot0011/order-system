package utils

import (
	"order-system/constants"
	"regexp"
)

// ValidateEmail validates email format using regex
func ValidateEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// IsValidOrderStatus checks if a status is valid
func IsValidOrderStatus(status string) bool {
	switch status {
	case constants.OrderStatusPending,
	     constants.OrderStatusConfirmed,
	     constants.OrderStatusPreparing,
	     constants.OrderStatusReady,
	     constants.OrderStatusDelivered,
	     constants.OrderStatusCompleted,
	     constants.OrderStatusCancelled:
		return true
	default:
		return false
	}
}

// IsValidFrontendOrderStatus checks if a frontend status is valid
func IsValidFrontendOrderStatus(status string) bool {
	switch status {
	case constants.FrontendOrderStatusActive,
	     constants.FrontendOrderStatusDelivered,
	     constants.FrontendOrderStatusPaid:
		return true
	default:
		return false
	}
}