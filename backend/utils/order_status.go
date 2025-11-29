package utils

import "order-system/constants"

// MapInternalStatusToFrontend maps internal order statuses to simplified frontend statuses
func MapInternalStatusToFrontend(internalStatus string) string {
	switch internalStatus {
	case constants.OrderStatusPending, constants.OrderStatusConfirmed, 
	     constants.OrderStatusPreparing, constants.OrderStatusReady:
		return constants.FrontendOrderStatusActive
	case constants.OrderStatusDelivered:
		return constants.FrontendOrderStatusDelivered
	case constants.OrderStatusCompleted:
		return constants.FrontendOrderStatusPaid
	default:
		return internalStatus // Return original if no mapping exists
	}
}

// MapFrontendStatusToInternal maps simplified frontend statuses to internal order statuses
func MapFrontendStatusToInternal(frontendStatus string) string {
	switch frontendStatus {
	case constants.FrontendOrderStatusActive:
		return constants.OrderStatusPending // Default internal status for active orders
	case constants.FrontendOrderStatusDelivered:
		return constants.OrderStatusDelivered
	case constants.FrontendOrderStatusPaid:
		return constants.OrderStatusCompleted
	default:
		return frontendStatus // Return original if no mapping exists
	}
}